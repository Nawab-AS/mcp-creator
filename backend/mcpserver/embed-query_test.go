package mcpserver

import (
	"reflect"
	"testing"

	ort "github.com/yalue/onnxruntime_go"
)

func TestBatchTokenDataPadsAndMasksQueries(t *testing.T) {
	inputIDs, attentionMask := batchTokenData([][]int64{{11, 12}, {21}}, 3, 0)

	if want := []int64{11, 12, 0, 21, 0, 0}; !reflect.DeepEqual(inputIDs, want) {
		t.Errorf("input IDs = %v, want %v", inputIDs, want)
	}
	if want := []int64{1, 1, 0, 1, 0, 0}; !reflect.DeepEqual(attentionMask, want) {
		t.Errorf("attention mask = %v, want %v", attentionMask, want)
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
		[]int64{1, 1, 0, 1, 0, 0},
	)
	if err != nil {
		t.Fatalf("collapseBatchEmbedding returned an error: %v", err)
	}

	want := [][]float32{{2, 20}, {5, 50}}
	if !reflect.DeepEqual(embeddings, want) {
		t.Errorf("embeddings = %v, want %v", embeddings, want)
	}
}
