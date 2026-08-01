package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Models struct{ ctx context.Context }

func (a *Models) Startup(ctx context.Context) {
	a.ctx = ctx

	// load (/save default) models save file
	configDir, err := GetConfigPath()
	if err != nil {
		fmt.Println("Error getting config path:", err)
		return
	}

	modelsFilePath := filepath.Join(configDir, "models.json")
	if _, err := os.Stat(modelsFilePath); os.IsNotExist(err) {
		// File does not exist, create it with default models
		if err := writeSaveFile(); err != nil {
			fmt.Println("Error writing default models to models.json:", err)
			return
		}
	} else if err == nil {
		// File exists, load it
		file, err := os.Open(modelsFilePath)
		if err != nil {
			fmt.Println("Error opening models.json:", err)
			return
		}
		defer file.Close()

		var loadedModels []Model
		if err := json.NewDecoder(file).Decode(&loadedModels); err != nil {
			fmt.Println("Error decoding models.json:", err)
			return
		}
		models = loadedModels
	}
}

var FS_MUTEX sync.Mutex

func writeSaveFile() error {
	FS_MUTEX.Lock()
	defer FS_MUTEX.Unlock()
	// Write `models` to `models.json` in the config directory
	configDir, err := GetConfigPath()
	if err != nil {
		fmt.Println("Error getting config path:", err)
		return err
	}

	modelsFilePath := filepath.Join(configDir, "models.json")
	file, err := os.Create(modelsFilePath)
	if err != nil {
		fmt.Println("Error creating models.json:", err)
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(models)
	if err != nil {
		fmt.Println("Error writing to models.json:", err)
		return err
	}
	return nil
}

type Model struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Installed   bool   `json:"installed"`
	Path        string `json:"path,omitempty"`

	OnnxURL      string  `json:"onnx_url,omitempty"`
	TokenizerURL string  `json:"tokenizer_url,omitempty"`
	SizeMB       float32 `json:"size_mb,omitempty"`
}

var models = []Model{
	{
		Name:        "MiniLM-L3-v2 (quantized)",
		Description: "The golden standard for most use cases. Outstanding efficiency for standard English text.",
		Installed:   false,

		OnnxURL:      "https://huggingface.co/Xenova/paraphrase-MiniLM-L3-v2/resolve/main/onnx/model_quantized.onnx",
		TokenizerURL: "https://huggingface.co/Xenova/paraphrase-MiniLM-L3-v2/resolve/main/tokenizer.json",
		SizeMB:       18.21,
	},
	{
		Name:        "Granite-embedding-small-english-r2-ONNX (quantized)",
		Description: "IBM's ultra-compact enterprise model. Highly optimized for handling code snippets and structured technical logs.",
		Installed:   false,

		OnnxURL:      "https://huggingface.co/onnx-community/granite-embedding-small-english-r2-ONNX/resolve/main/onnx/model_quantized.onnx",
		TokenizerURL: "https://huggingface.co/onnx-community/granite-embedding-small-english-r2-ONNX/resolve/main/tokenizer.json",
		SizeMB:       2.73,
	},
	{
		Name:        "Snowflake-arctic-embed-xs (quantized)",
		Description: "Top-tier accuracy for complex documents. Specifically engineered to inderstand deep topics from lengthy files while maintaining a tiny footprint.",
		Installed:   false,

		OnnxURL:      "https://huggingface.co/Snowflake/snowflake-arctic-embed-xs/resolve/main/onnx/model_quantized.onnx",
		TokenizerURL: "https://huggingface.co/Snowflake/snowflake-arctic-embed-xs/resolve/main/tokenizer.json",
		SizeMB:       23.71,
	},
	{
		Name:        "Cescofors75/baco-embeddings",
		Description: "Ultra-lightweight extreme efficiency model. Sub-millisecond execution times designed specifically for weak hardware.",
		Installed:   false,

		OnnxURL:      "https://huggingface.co/Cescofors75/baco-embeddings/resolve/main/onnx/fine_tuned_model.onnx",
		TokenizerURL: "https://huggingface.co/Cescofors75/baco-embeddings/resolve/main/tokenizer.json",
		SizeMB:       1.13,
	},
}

func (a *Models) DownloadModel(modelName string) {
	fmt.Println("Downloading model:", modelName)

	if modelName == "" {
		return
	}

	var model *Model
	for i := range models {
		if models[i].Name == modelName {
			model = &models[i]
			break
		}
	}

	if model == nil {
		fmt.Println("Model not found:", modelName)
		return
	}

	if model.Installed && model.Path != "" {
		fmt.Println("Model already installed:", modelName)
		return
	}

	go func() {
		runtime.EventsEmit(a.ctx, "model-download-started", modelName)

		modelDirectory, err := ModelStorageDirectory(modelName)
		if err != nil {
			fmt.Printf("Unable to prepare model directory for %q: %v\n", modelName, err)
			return
		}

		stagingDirectory := modelDirectory + ".partial"
		if err := os.RemoveAll(stagingDirectory); err != nil {
			fmt.Printf("Unable to clear partial download for %q: %v\n", modelName, err)
			return
		}
		if err := os.MkdirAll(stagingDirectory, 0o755); err != nil {
			fmt.Printf("Unable to create model directory for %q: %v\n", modelName, err)
			return
		}

		download := func(sourceURL string, filename string, baseProgress float64, weight float64) error {
			return DownloadFile(sourceURL, filepath.Join(stagingDirectory, filename), func(progress float64) {
				runtime.EventsEmit(a.ctx, "model-download-progress", modelName, baseProgress+(progress*weight))
			})
		}

		if err := download(model.OnnxURL, "model.onnx", 0, 0.95); err != nil {
			fmt.Printf("Unable to download model %q: %v\n", modelName, err)
			_ = os.RemoveAll(stagingDirectory)
			return
		}
		if err := download(model.TokenizerURL, "tokenizer.json", 0.95, 0.05); err != nil {
			fmt.Printf("Unable to download tokenizer for %q: %v\n", modelName, err)
			_ = os.RemoveAll(stagingDirectory)
			return
		}

		if err := os.RemoveAll(modelDirectory); err != nil {
			fmt.Printf("Unable to replace model %q: %v\n", modelName, err)
			_ = os.RemoveAll(stagingDirectory)
			return
		}
		if err := os.Rename(stagingDirectory, modelDirectory); err != nil {
			fmt.Printf("Unable to finalize model %q: %v\n", modelName, err)
			_ = os.RemoveAll(stagingDirectory)
			return
		}

		writeSaveFile()
		model.Installed = true
		model.Path = filepath.Join(modelDirectory, "model.onnx")
		runtime.EventsEmit(a.ctx, "model-download-completed", modelName)
	}()
}

func (a *Models) DeleteModel(modelName string) {
	fmt.Println("Deleting model:", modelName)
	if modelName == "" {
		return
	}

	var model *Model
	for i := range models {
		if models[i].Name == modelName {
			model = &models[i]
			break
		}
	}

	if model == nil {
		fmt.Println("Model not found:", modelName)
		return
	}

	if !model.Installed && model.Path == "" {
		fmt.Println("Model already not installed:", modelName)
		return
	}

	go func() {
		runtime.EventsEmit(a.ctx, "model-delete-started", modelName)

		modelDirectory, err := ModelStorageDirectory(modelName)
		if err != nil {
			fmt.Printf("Unable to locate model directory for %q: %v\n", modelName, err)
			return
		}

		deleteDone := make(chan error, 1)
		go func() {
			deleteDone <- os.RemoveAll(modelDirectory)
		}()

		ticker := time.NewTicker(progressUpdateInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				runtime.EventsEmit(a.ctx, "model-delete-progress", modelName, 0)
			case err := <-deleteDone:
				if err != nil {
					fmt.Printf("Unable to delete model %q: %v\n", modelName, err)
					return
				}

				model.Installed = false
				model.Path = ""
				writeSaveFile()
				// runtime.EventsEmit(a.ctx, "model-delete-progress", modelName, 1)
				runtime.EventsEmit(a.ctx, "model-delete-completed", modelName)
				return
			}
		}
	}()
}

// expose to entire package `backend`
func (a *Models) GetModels() []Model { return GetModels() }

func GetModels() []Model {
	FS_MUTEX.Lock()
	defer FS_MUTEX.Unlock()
	return models
}
