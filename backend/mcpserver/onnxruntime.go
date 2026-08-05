package mcpserver

// cross-platform is not for now. rn its only intel-based macs with versions 12

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	ort "github.com/yalue/onnxruntime_go"
)

var common CommonFuncs

type CommonFuncs interface {
	GetConfigPath() (string, error)
	FsExists(path string, isDir bool) (bool, error)
	DownloadFile(sourceURL string, targetPath string, onProgress func(float32)) error
	EmitEvent(eventName string, data ...interface{})
	ModelStorageDirectory(modelName string) (string, error)
	ProjectDBPath(projectName string) (string, string, error)
	Hash(input string) string
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
		fmt.Println("Error determining onnxruntime binary name:", err)
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
	e := ort.InitializeEnvironment()
	if e != nil {
		fmt.Printf("Error initializing the onnxruntime library: %v\n", e)
		return e
	}
	end := time.Now().UnixNano()
	fmt.Printf("Time taken to load environment: %d ms\n", (end-start)/1e6)
	return nil
}

func getOnnxBinaryName() (string, error) {
	supportedLibraries := []string{
		"onnxruntime-linux-aarch64.so",
		"onnxruntime-linux-x64.so",
		"onnxruntime-osx-arm64.dylib",
		"onnxruntime-osx-x86_64.dylib",
		"onnxruntime-win-arm.dll",
		"onnxruntime-win-arm64.dll",
		"onnxruntime-win-x64.dll",
		"onnxruntime-win-x86.dll",
	}
	ending, os, arch := "", "", runtime.GOARCH
	switch runtime.GOOS {
	case "darwin":
		ending = "dylib"
		os = "osx"
		if arch == "amd64" {
			arch = "x86_64"
		}
	case "linux":
		ending = "so"
		os = "linux"
	case "windows":
		ending = "dll"
		os = "win"
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	name := fmt.Sprintf("onnxruntime-%s-%s.%s", os, arch, ending)
	for _, lib := range supportedLibraries {
		if lib == name {
			return name, nil
		}
	}

	return "", fmt.Errorf("unsupported architecture: %s", name)
}

func downloadOnnxRuntime(targetPath string) error {
	common.EmitEvent("started", "onnxruntime-download")
	onnxBinaryName, err := getOnnxBinaryName()
	if err != nil {
		fmt.Println("Error determining onnxruntime binary name:", err)
		return err
	}

	downloadURL := fmt.Sprintf("https://github.com/Nawab-AS/legacy-onnx-runtimes/raw/refs/heads/main/%s", onnxBinaryName)

	err = common.DownloadFile(downloadURL, targetPath, func(progress float32) {
		common.EmitEvent("progress", "onnxruntime-download", progress)
	})

	if err != nil {
		common.EmitEvent("error", "onnxruntime-download", "Network error")
		return err
	}

	common.EmitEvent("completed", "onnxruntime-download")

	return nil
}
