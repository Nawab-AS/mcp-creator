package files

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// download.go - File download utility with progress reporting and file locking
//
// Example usage:
//
//	func someFunction() {
//		files := &Files{}
//		err := files.DownloadFile("https://example.com/file.zip", "/absolute/path/to/file.zip", func(percentage float32) {
//			fmt.Printf("Download progress: %.2f%%\n", percentage)
//		})
//		if err != nil {
//			fmt.Printf("Error downloading file: %v\n", err)
//		}
//	}

var downloadClient = &http.Client{Timeout: 10 * time.Minute}

type countingWriter struct {
	writer  io.Writer
	written atomic.Int64
}

func (w *countingWriter) Write(data []byte) (int, error) {
	n, err := w.writer.Write(data)
	w.written.Add(int64(n))
	return n, err
}

// DownloadFile synchronously downloads a file from sourceURL to targetPath, reporting progress via onProgressRaw callback.
func (f *Files) DownloadFile(sourceURL string, targetPath string, onProgressRaw func(percentage float32)) error {
	progressUpdate := func(percentage float32) {
		if onProgressRaw != nil {
			onProgressRaw(percentage)
		}
	}

	// aquire mutex lock
	mutex := f.Mutex(targetPath).Lock()
	defer mutex.Unlock()

	// create HTTP request
	request, err := http.NewRequest(http.MethodGet, sourceURL, nil)
	if err != nil {
		return fmt.Errorf("error creating download request: %w", err)
	}

	// send request
	response, err := downloadClient.Do(request)
	if err != nil {
		return fmt.Errorf("error sending download request for %s: %w", sourceURL, err)
	}
	defer response.Body.Close()

	// check for HTTP errors
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("error downloading %s: unexpected status %s", sourceURL, response.Status)
	}

	// create temporary file for download
	tempFileName := filepath.Base(targetPath) + ".tmp-*"
	temporaryFile, err := os.CreateTemp(filepath.Dir(targetPath), tempFileName)
	if err != nil {
		return fmt.Errorf("error creating temporary file for %s: %w", targetPath, err)
	}
	temporaryPath := temporaryFile.Name()
	defer os.Remove(temporaryPath)
	defer temporaryFile.Close() // panic fallback

	// write to temporary file while reporting progress
	writer := &countingWriter{writer: temporaryFile}
	stopProgressReport := make(chan struct{})
	progressReportStopped := make(chan struct{})

	// ensure progress reporting stops and channel is closed when function exits
	var stopProgressReportOnce sync.Once
	stopProgressReporter := func() {
		stopProgressReportOnce.Do(func() {
			close(stopProgressReport)
			<-progressReportStopped
		})
	}
	defer stopProgressReporter() // panic fallback

	// reports progress asynchronously while downloading synchronously
	go func() {
		defer close(progressReportStopped)
		ticker := time.NewTicker(progressUpdateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				progress := float32(writer.written.Load()) / float32(response.ContentLength) * 100

				// only report progress if we know the total size, but not 100% (finalizing temp file)
				if response.ContentLength > 0 && progress <= 99 {
					progressUpdate(progress)
				}
			case <-stopProgressReport:
				return // stop progress reporting and closes progressDone
			}
		}
	}()

	// stream response to temporary file (blocking call)
	if _, err := io.Copy(writer, response.Body); err != nil {
		return fmt.Errorf("error downloading %s: %w", sourceURL, err)
	}

	stopProgressReporter()

	if err := temporaryFile.Close(); err != nil {
		return fmt.Errorf("error closing temporary file: %w", err)
	}

	if err := os.Rename(temporaryPath, targetPath); err != nil {
		return fmt.Errorf("error moving temporary file to '%s': %w", targetPath, err)
	}

	// final progress update to 100%
	progressUpdate(100)
	return nil
}

func (f *Files) ParallelDownloadFiles(downloads []struct {
	SourceURL  string
	TargetPath string
}, onProgressRaw func(percentage float32)) error {
	var numDownloads = len(downloads)
	if len(downloads) == 0 {
		return fmt.Errorf("no source URLs provided for download")
	}

	onProgress := func(percentage float32) {
		if onProgressRaw != nil {
			onProgressRaw(percentage)
		}
	}

	currentSubProgress := make([]float32, numDownloads)
	var progressMu sync.Mutex
	averageProgress := func() float32 {
		progressMu.Lock()
		defer progressMu.Unlock()

		var sum float32
		for _, p := range currentSubProgress {
			sum += p
		}
		return sum / float32(numDownloads)
	}

	var wg sync.WaitGroup
	downloadsFinished := make(chan struct{})
	go func() {
		wg.Wait()
		close(downloadsFinished)
	}()

	var failed atomic.Bool
	for i := 0; i < numDownloads; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := f.DownloadFile(downloads[i].SourceURL, downloads[i].TargetPath, func(percentage float32) {
				progressMu.Lock()
				currentSubProgress[i] = percentage
				progressMu.Unlock()
			})

			if err != nil {
				fmt.Printf("Error downloading %s: %v\n", downloads[i].SourceURL, err)
				failed.Store(true)
			}
		}(i)
	}

	// progress updating
	ticker := time.NewTicker(progressUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			onProgress(averageProgress())

		case <-downloadsFinished:
			if failed.Load() { // if any download failed, remove all downloads (transactional)
				for i := 0; i < numDownloads; i++ {
					// Remove every target after all workers finish so no download can recreate it.
					_ = os.Remove(downloads[i].TargetPath)
				}
				return fmt.Errorf("one or more downloads failed")
			}
			onProgress(100)
			return nil
		}
	}
}
