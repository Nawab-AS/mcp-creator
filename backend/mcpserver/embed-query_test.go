package mcpserver

import (
	"math"
	"path/filepath"
	"reflect"
	"testing"

	ort "github.com/yalue/onnxruntime_go"
)

func TestBatchTokenDataPadsAndMasksQueries(t *testing.T) {
	inputIDs, attentionMask := batchTokenData([][]int32{{11, 12}, {21}}, 3, 0)

	if want := []int32{11, 12, 0, 21, 0, 0}; !reflect.DeepEqual(inputIDs, want) {
		t.Errorf("input IDs = %v, want %v", inputIDs, want)
	}
	if want := []int32{1, 1, 0, 1, 0, 0}; !reflect.DeepEqual(attentionMask, want) {
		t.Errorf("attention mask = %v, want %v", attentionMask, want)
	}
}

func TestVectorDatabasePath(t *testing.T) {
	model := &modelData{ModelPath: filepath.Join(t.TempDir(), "model.onnx")}

	path, err := vectorDatabasePath(model)
	if err != nil {
		t.Fatalf("vectorDatabasePath returned an error: %v", err)
	}

	want := filepath.Join(filepath.Dir(model.ModelPath), vectorDatabaseFilename)
	if path != want {
		t.Errorf("vector database path = %q, want %q", path, want)
	}
}

func TestVectorDatabasePathRequiresModelPath(t *testing.T) {
	for _, model := range []*modelData{nil, {}} {
		if _, err := vectorDatabasePath(model); err == nil {
			t.Errorf("vectorDatabasePath(%v) returned no error", model)
		}
	}
}

func TestModelVectorDimension(t *testing.T) {
	model := &modelData{Output: []ort.InputOutputInfo{{Dimensions: []int64{1, -1, 384}}}}

	dimension, err := modelVectorDimension(model)
	if err != nil {
		t.Fatalf("modelVectorDimension returned an error: %v", err)
	}
	if dimension != 384 {
		t.Errorf("vector dimension = %d, want 384", dimension)
	}
}

func TestNormalizeVector(t *testing.T) {
	input := []float32{3, 4}
	normalized, err := normalizeVector(input)
	if err != nil {
		t.Fatalf("normalizeVector returned an error: %v", err)
	}

	if math.Abs(float64(normalized[0])-0.6) > 1e-6 || math.Abs(float64(normalized[1])-0.8) > 1e-6 {
		t.Errorf("normalized vector = %v, want [0.6 0.8]", normalized)
	}
	if !reflect.DeepEqual(input, []float32{3, 4}) {
		t.Errorf("normalizeVector mutated input: %v", input)
	}
}

func TestVectorDBKNN(t *testing.T) {
	model := &modelData{
		ModelPath: filepath.Join(t.TempDir(), "model.onnx"),
		Output:    []ort.InputOutputInfo{{Dimensions: []int64{1, 2}}},
	}
	vdb, err := NewVectorDB(model)
	if err != nil {
		t.Fatalf("NewVectorDB returned an error: %v", err)
	}
	defer vdb.Close()

	for _, row := range []vectorDataRow{
		{Filename: "near.txt", text: "near", Vector: []float32{0, 1}},
		{Filename: "far.txt", text: "far", Vector: []float32{2, 2}},
	} {
		if err := vdb.InsertVector(row); err != nil {
			t.Fatalf("InsertVector returned an error: %v", err)
		}
	}

	results, err := vdb.HybridSearch([]float32{0, 0}, 2)
	if err != nil {
		t.Fatalf("HybridSearch returned an error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("HybridSearch returned %d results, want 2", len(results))
	}
	if results[0].Chunk_ID != 1 || results[0].text != "near" {
		t.Errorf("nearest result = %+v, want chunk 1 with near text", results[0])
	}
}

func TestCollapseBatchEmbeddingExcludesPadding(t *testing.T) {
	data := []float32{
		1, 10, 3, 30, 100, 1000,
		5, 50, 200, 2000, 300, 3000,
	}

	embeddings, err := collapseBatchEmbedding(
		data,
		ort.NewShape(2, 3, 2),
		2,
		[]int32{1, 1, 0, 1, 0, 0},
	)
	if err != nil {
		t.Fatalf("collapseBatchEmbedding returned an error: %v", err)
	}

	want := [][]float32{{2, 20}, {5, 50}}
	if !reflect.DeepEqual(embeddings, want) {
		t.Errorf("embeddings = %v, want %v", embeddings, want)
	}
}
