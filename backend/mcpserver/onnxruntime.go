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

	envLoaded := time.Now().UnixNano()
	model, err := loadModel("MiniLM-L3-v2 (quantized)")
	if err != nil {
		fmt.Println("Error loading model:", err)
		return err
	}

	fmt.Println("Model loaded successfully")
	queries := []string{
		"This is a test to see if the embedding is working or not",
		"Nuclear fission is baically an atom splitting, releasing energy in the process",
		"Apple pie is a delicious dessert usually served with whipped cream",
		"One more test to see if the embedding is working or not",
		"A banana boat is a type of dessert made with bananas and ice cream",
		"Really final test to see if the embedding is working or not",
		"The distance between the Earth and the Moon is about 238,855 miles (384,400 kilometers)",
		"Okay, this is the last test to see if the embedding is working or not",
		"And lastly, this is the final sentence BUT its very long and should be truncated to see if the embedding is working or not",
	}

	modelLoaded := time.Now().UnixNano()
	embeddings, err := model.BatchEmbedQueries(queries)
	if err != nil {
		fmt.Println("Error embedding query:", err)
		return err
	}

	for i, embedding := range embeddings {
		fmt.Printf("Embedding for query %d: len=%d, text=%s\n", i, len(embedding), queries[i])
	}
	embeddingsGenerated := time.Now().UnixNano()

	// vectorDB test
	vdb, err := NewVectorDB(&modelData{
		Name:      "MiniLM-L3-v2 (quantized)",
		ModelPath: model.ModelPath,
		Output:    model.Output,
	})
	if err != nil {
		fmt.Println("Error creating vector database:", err)
		return err
	}
	defer vdb.Close()
	vectorDBCreated := time.Now().UnixNano()

	for i, row := range embeddings {
		if err := vdb.InsertVector(vectorDataRow{
			Chunk_ID: int32(i),
			Filename: "test.txt",
			text:     queries[i],
			Vector:   row,
		}); err != nil {
			fmt.Println("Error inserting vector:", err)
			return err
		}
	}
	vectorsInserted := time.Now().UnixNano()

	query, err := model.EmbedQuery("Apple pie is a yummy dessert")
	if err != nil {
		fmt.Println("Error embedding query:", err)
		return err
	}

	results, err := vdb.HybridSearch(query, 5)
	if err != nil {
		fmt.Println("Error performing hybrid search:", err)
		return err
	}

	fmt.Println("Hybrid search results:")
	for _, result := range results {
		fmt.Printf("  Chunk_ID: %d  Text: %s\n", result.Chunk_ID, result.text)
	}
	searchingFinished := time.Now().UnixNano()

	vdb.WriteToDisk()
	vdb.Close()
	end := time.Now().UnixNano()

	fmt.Printf("\n\n\nTime taken to load environment: %d ms\n", (envLoaded-start)/1e6)
	fmt.Printf("Time taken to load model: %d ms\n", (modelLoaded-envLoaded)/1e6)
	fmt.Printf("Time taken to generate embeddings: %d ms\n", (embeddingsGenerated-modelLoaded)/1e6)
	fmt.Printf("Time taken to create vector database: %d ms\n", (vectorDBCreated-embeddingsGenerated)/1e6)
	fmt.Printf("Time taken to insert vectors: %d ms\n", (vectorsInserted-vectorDBCreated)/1e6)
	fmt.Printf("Time taken to query vector database: %d ms\n", (searchingFinished-vectorsInserted)/1e6)
	fmt.Printf("Time taken to write vector database to disk: %d ms\n", (end-searchingFinished)/1e6)
	fmt.Printf("Total time taken: %d ms\n", (end-start)/1e6)

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
