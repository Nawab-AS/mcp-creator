package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Models struct{}

func (a *Models) Startup(ctx context.Context) {
	// load (/save default) models save file
	modelsFilePath := Utils.Files.GetFileStoragePath("models.json")
	if _, err := os.Stat(modelsFilePath); os.IsNotExist(err) {
		// File does not exist, create it with default models
		Utils.Files.WriteStructToFile(models, "models.json")

	} else if err == nil { // File exists, load it
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

	refreshInstalledModelState()
}

func refreshInstalledModelState() {
	for i := range models {
		modelDirectory := Utils.Files.ModelStorageDirectory(models[i].Name)

		modelFilePath := filepath.Join(modelDirectory, "model.onnx")
		if _, err := os.Stat(modelFilePath); err == nil {
			models[i].Installed = true
			models[i].Path = modelFilePath
			continue
		}

		if !models[i].Installed {
			models[i].Path = ""
		}
	}
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
		Name:        "mxbai-embed-xsmall-v1 (quantized)",
		Description: "A compact, efficient embedding model for semantic search with low resource usage.",
		Installed:   false,

		OnnxURL:      "https://huggingface.co/mixedbread-ai/mxbai-embed-xsmall-v1/resolve/main/onnx/model_quantized.onnx",
		TokenizerURL: "https://huggingface.co/mixedbread-ai/mxbai-embed-xsmall-v1/resolve/main/tokenizer.json",
		SizeMB:       25.11,
	},
	{
		Name:        "Snowflake-arctic-embed-xs (quantized)",
		Description: "Top-tier accuracy for complex documents. Specifically engineered to inderstand deep topics from lengthy files while maintaining a tiny footprint.",
		Installed:   false,

		OnnxURL:      "https://huggingface.co/Snowflake/snowflake-arctic-embed-xs/resolve/main/onnx/model_quantized.onnx",
		TokenizerURL: "https://huggingface.co/Snowflake/snowflake-arctic-embed-xs/resolve/main/tokenizer.json",
		SizeMB:       23.71,
	},
}

func (a *Models) DownloadModel(modelName string) {
	fmt.Println("Downloading model:", modelName)

	// find the model by name
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
	} else if model.Installed && model.Path != "" {
		fmt.Println("Model already installed:", modelName)
		return
	}

	// start download
	Events.Emit("started", "model-download", modelName)
	modelDirectory := Utils.Files.ModelStorageDirectory(modelName)

	err := Utils.Files.ParallelDownloadFiles([]struct {
		SourceURL  string
		TargetPath string
	}{
		{model.OnnxURL, filepath.Join(modelDirectory, "model.onnx")},
		{model.TokenizerURL, filepath.Join(modelDirectory, "tokenizer.json")},
	}, func(percentage float32) {
		Events.Emit("progress", "model-download", modelName, percentage)
	})

	if err != nil {
		fmt.Printf("Unable to parallel download: %v", err)
		Events.Emit("error", "model-download", modelName, "Network error")
		return
	}
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

	Events.Emit("started", "model-delete", modelName)

	modelDirectory := Utils.Files.ModelStorageDirectory(modelName)

	if !Utils.Files.IsDirWritable(modelDirectory) {
		fmt.Printf("Unable to access model directory for %q", modelName)
		Events.Emit("error", "model-delete", modelName, "File not found")
		return
	}

	if err := os.RemoveAll(modelDirectory); err != nil {
		fmt.Printf("Unable to delete model directory for %q: %v", modelName, err)
		Events.Emit("error", "model-delete", modelName, "Permission denied")
		return
	}

	model.Installed = false
	model.Path = ""
	Utils.Files.WriteStructToFile(models, "models.json")
	Events.Emit("completed", "model-delete", modelName)
}
