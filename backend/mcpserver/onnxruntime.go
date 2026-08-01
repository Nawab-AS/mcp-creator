package mcpserver

// cross-platform is not for now. rn its only intel-based macs with versions 12 and below

import (
	"fmt"
	"path/filepath"

	ort "github.com/yalue/onnxruntime_go"
)

var common CommonFuncs

type CommonFuncs interface {
	GetConfigPath() (string, error)
	FsExists(path string, isDir bool) (bool, error)
	DownloadFile(sourceURL string, targetPath string, onProgress func(float64)) error
	EmitEvent(eventName string, data ...interface{})
}

func OnnxStartup(commonFuncs1 CommonFuncs) error {
	common = commonFuncs1
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
		common.EmitEvent("onnxruntime-download-started", "ONNX Runtime library")
		if err := downloadOnnxRuntime(onnxBinaryPath); err != nil {
			return err
		}
		common.EmitEvent("onnxruntime-download-completed", "ONNX Runtime library")
	}

	ort.SetSharedLibraryPath(onnxBinaryPath)
	e := ort.InitializeEnvironment()
	if e != nil {
		fmt.Printf("Error initializing the onnxruntime library: %v\n", e)
		return e
	}

	return nil
}

func getOnnxBinaryName() (string, error) {
	return "onnxruntime_osx_legacy_amd64.dylib", nil
	// if runtime.GOOS == "windows" {
	// 	if runtime.GOARCH == "amd64" {
	// 		return "onnxruntime.dll", nil
	// 	}
	// }
	// if runtime.GOOS == "darwin" {
	// 	if runtime.GOARCH == "arm64" {
	// 		return "onnxruntime_arm64.dylib", nil
	// 	}
	// 	if runtime.GOARCH == "amd64" {
	// 		return "onnxruntime_amd64.dylib", nil
	// 	}
	// }
	// if runtime.GOOS == "linux" {
	// 	if runtime.GOARCH == "arm64" {
	// 		return "onnxruntime_arm64.so", nil
	// 	}
	// 	return "onnxruntime.so", nil
	// }
	// fmt.Printf("Unable to determine a path to the onnxruntime shared library"+
	// 	" for OS \"%s\" and architecture \"%s\".\n", runtime.GOOS,
	// 	runtime.GOARCH)
	// return "", errors.New("unable to determine a path to the onnxruntime shared library")
}

func downloadOnnxRuntime(targetPath string) error {
	// onnxBinaryName, err := getOnnxBinaryName()
	// if err != nil {
	// 	fmt.Println("Error determining onnxruntime binary name:", err)
	// 	return err
	// }

	// downloadURL := fmt.Sprintf("https://github.com/yalue/onnxruntime_go_examples/raw/refs/heads/master/third_party/%s", onnxBinaryName)
	downloadURL := "https://github.com/Nawab-AS/osx-legacy-onnxruntime/raw/refs/heads/main/onnxruntime_osx_legacy_x84_64.dylib"

	return common.DownloadFile(downloadURL, targetPath, func(progress float64) {
		common.EmitEvent("onnxruntime-download-progress", "ONNX Runtime library", progress)
	})
}
