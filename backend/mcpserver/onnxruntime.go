package mcpserver

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	ort "github.com/yalue/onnxruntime_go"
)

var common CommonFuncs

const onnxRuntimeVersion = "1.30.0"

type CommonFuncs interface {
	GetConfigPath() (string, error)
	FsExists(path string, isDir bool) (bool, error)
	DownloadFile(sourceURL string, targetPath string, onProgress func(float32)) error
	EmitEvent(eventName string, data ...interface{})
	ModelStorageDirectory(modelName string) (string, error)
	ProjectDBPath(projectName string) (string, string, error)
	Hash(input string) string
	ShowInfoDialog(message string)
	Uninstall()
}

func OnnxStartup(commonFuncs CommonFuncs) error {
	common = commonFuncs
	// load/save onnxruntime shared library
	configDir, err := common.GetConfigPath()
	if err != nil {
		fmt.Println("Error getting config path:", err)
		return err
	}

	onnxBinaryName, err := getOnnxBinaryName()
	if err != nil {
		common.ShowInfoDialog(err.Error())
		common.Uninstall()
		return err
	}

	onnxBinaryPath := filepath.Join(configDir, onnxBinaryName)

	if fsExists, err := common.FsExists(onnxBinaryPath, false); err != nil {
		fmt.Println("Error checking if onnxruntime binary exists:", err)
		return err
	} else if !fsExists {
		fmt.Println("onnxruntime binary does not exist, downloading...")
		if err := downloadOnnxRuntime(onnxBinaryPath); err != nil {
			return err
		}
	}

	start := time.Now().UnixNano()
	ort.SetSharedLibraryPath(onnxBinaryPath)
	if err := ort.InitializeEnvironment(); err != nil {
		initErr := fmt.Errorf("error initializing the ONNX Runtime library: %w", err)
		fmt.Println(initErr)
		common.ShowInfoDialog(initErr.Error())
		common.Uninstall()
		return initErr
	}
	end := time.Now().UnixNano()
	fmt.Printf("Time taken to load environment: %d ms\n", (end-start)/1e6)
	return nil
}

func getOnnxBinaryName() (string, error) {
	switch {
	case runtime.GOOS == "windows" && runtime.GOARCH == "amd64":
		return "onnxruntime-win-x64-" + onnxRuntimeVersion + ".dll", nil
	case runtime.GOOS == "darwin" && runtime.GOARCH == "arm64":
		return "onnxruntime-osx-arm64-" + onnxRuntimeVersion + ".dylib", nil
	default:
		return "", fmt.Errorf("unsupported platform %s/%s; supported platforms are windows/amd64 and darwin/arm64", runtime.GOOS, runtime.GOARCH)
	}
}

func downloadOnnxRuntime(targetPath string) error {
	common.EmitEvent("started", "onnxruntime-download")
	onnxBinaryName, err := getOnnxBinaryName()
	if err != nil {
		return err
	}

	archiveName := "onnxruntime-win-x64-" + onnxRuntimeVersion + ".zip"
	if runtime.GOOS == "darwin" {
		archiveName = "onnxruntime-osx-arm64-" + onnxRuntimeVersion + ".tgz"
	}
	downloadURL := fmt.Sprintf("https://github.com/microsoft/onnxruntime/releases/download/v%s/%s", onnxRuntimeVersion, archiveName)
	archivePath := targetPath + ".archive"
	defer os.Remove(archivePath)

	err = common.DownloadFile(downloadURL, archivePath, func(progress float32) {
		common.EmitEvent("progress", "onnxruntime-download", progress)
	})

	if err != nil {
		common.EmitEvent("error", "onnxruntime-download", "Network error")
		return err
	}
	if err := extractRuntime(archivePath, targetPath, onnxBinaryName); err != nil {
		common.EmitEvent("error", "onnxruntime-download", "Invalid runtime archive")
		return err
	}

	common.EmitEvent("completed", "onnxruntime-download")

	return nil
}

func extractRuntime(archivePath string, targetPath string, binaryName string) error {
	if runtime.GOOS == "windows" {
		return extractWindowsRuntime(archivePath, targetPath)
	}
	return extractDarwinRuntime(archivePath, targetPath, binaryName)
}

func extractWindowsRuntime(archivePath string, targetPath string) error {
	archive, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("open ONNX Runtime archive: %w", err)
	}
	defer archive.Close()

	for _, file := range archive.File {
		name := filepath.ToSlash(file.Name)
		if !strings.HasSuffix(name, "/lib/onnxruntime.dll") && !strings.HasSuffix(name, "/lib/onnxruntime_providers_shared.dll") {
			continue
		}
		outputName := filepath.Base(name)
		if outputName == "onnxruntime.dll" {
			outputName = filepath.Base(targetPath)
		}
		if err := extractZipFile(file, filepath.Join(filepath.Dir(targetPath), outputName)); err != nil {
			return err
		}
	}
	if _, err := os.Stat(targetPath); err != nil {
		return fmt.Errorf("ONNX Runtime DLL was not found in archive: %w", err)
	}
	return nil
}

func extractZipFile(file *zip.File, targetPath string) error {
	input, err := file.Open()
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", targetPath, err)
	}
	if _, err := io.Copy(output, input); err != nil {
		output.Close()
		return fmt.Errorf("extract %s: %w", targetPath, err)
	}
	return output.Close()
}

func extractDarwinRuntime(archivePath string, targetPath string, binaryName string) error {
	archiveFile, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer archiveFile.Close()
	gzipReader, err := gzip.NewReader(archiveFile)
	if err != nil {
		return fmt.Errorf("open ONNX Runtime archive: %w", err)
	}
	defer gzipReader.Close()

	reader := tar.NewReader(gzipReader)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read ONNX Runtime archive: %w", err)
		}
		if !strings.HasSuffix(header.Name, "/lib/libonnxruntime.1."+onnxRuntimeVersion+".dylib") {
			continue
		}
		output, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(output, reader); err != nil {
			output.Close()
			return err
		}
		if err := output.Close(); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("%s was not found in archive", binaryName)
}
