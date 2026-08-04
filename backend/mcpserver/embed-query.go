package mcpserver

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sugarme/tokenizer"
	"github.com/sugarme/tokenizer/pretrained"

	ort "github.com/yalue/onnxruntime_go"
)

type Tokenizer struct {
	Tokenizer          *tokenizer.Tokenizer
	MaxSequenceLength int
}

type modelData struct {
	Name      string
	ModelPath string
	Tokenizer Tokenizer
	Input     []ort.InputOutputInfo
	Output    []ort.InputOutputInfo
}

func LoadTokenizer(tokenizer_path string) (*Tokenizer, error) {
	t, err := pretrained.FromFile(tokenizer_path)
	if err != nil {
		return nil, err
	}
	maxSequenceLength := 0
	if truncation := t.GetTruncation(); truncation != nil {
		maxSequenceLength = truncation.MaxLength
	}
	t.WithTruncation(nil)
	return &Tokenizer{
		Tokenizer:          t,
		MaxSequenceLength: maxSequenceLength,
	}, nil
}

func (t *Tokenizer) Encode(text string) ([]int32, error) {
	encoding, err := t.Tokenizer.EncodeSingle(text, true)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("%-10s: %q\n", "Tokens", encoding.Tokens)
	// fmt.Printf("%-10s: %v\n", "Ids", encoding.Ids)

	ids := make([]int32, len(encoding.Ids))
	for i, id := range encoding.Ids {
		ids[i] = int32(id)
	}

	return ids, nil
}

func (t *Tokenizer) Decode(ids []int32) string {
	decoded := make([]int, len(ids))
	for index, id := range ids {
		decoded[index] = int(id)
	}
	return t.Tokenizer.Decode(decoded, true)
}

func (t *Tokenizer) EncodeBatch(queries []string) ([][]int32, error) {
	// encodes a batch of queries into token IDs using the tokenizer
	batchInput := make([]tokenizer.EncodeInput, len(queries))
	for i, s := range queries {
		batchInput[i] = tokenizer.NewSingleEncodeInput(tokenizer.NewInputSequence(s))
	}

	encodings, err := t.Tokenizer.EncodeBatch(batchInput, true)
	if err != nil {
		return nil, err
	}

	// convert encodings to [][]int32
	batchIds := make([][]int32, len(encodings))
	for i, encoding := range encodings {
		ids := make([]int32, len(encoding.Ids))
		for j, id := range encoding.Ids {
			ids[j] = int32(id)
		}
		batchIds[i] = ids
	}

	return batchIds, nil
}

func loadModel(modelName string) (modelData, error) {
	modelDir, err := common.ModelStorageDirectory(modelName)
	if err != nil {
		fmt.Println("Error loading model directory:", err)
		return modelData{}, err
	}

	modelPath := filepath.Join(modelDir, "model.onnx")
	inputs, outputs, err := ort.GetInputOutputInfo(modelPath)
	if err != nil {
		fmt.Println("Error loading model I/O info:", err)
		return modelData{}, err
	}

	tokenizer, err := LoadTokenizer(filepath.Join(modelDir, "tokenizer.json"))
	if err != nil {
		fmt.Println("Error loading tokenizer:", err)
		return modelData{}, err
	}

	return modelData{
		Name:      modelName,
		ModelPath: modelPath,
		Tokenizer: *tokenizer,
		Input:     inputs,
		Output:    outputs,
	}, nil
}

func (m *modelData) Tokenize(text string) ([]int32, error) {
	return m.Tokenizer.Encode(text)
}

func (m *modelData) EmbedQuery(query string) ([]float32, error) {
	tokenIds, err := m.Tokenize(query)
	if err != nil {
		return nil, err
	}

	inputs := make([]ort.ArbitraryTensor, len(m.Input))
	defer destroyTensors(inputs)
	for i, inputInfo := range m.Input {
		value, err := buildInputTensor(inputInfo, tokenIds)
		if err != nil {
			return nil, err
		}
		inputs[i] = value
	}

	outputs, err := m.run(inputs)
	if err != nil {
		return nil, err
	}
	defer destroyTensors(outputs)

	return extractEmbedding(outputs)
}

func (m *modelData) BatchEmbedQueries(queries []string) ([][]float32, error) {
	if len(queries) == 0 {
		return [][]float32{}, nil
	}

	tokens, err := m.Tokenizer.EncodeBatch(queries)
	if err != nil {
		return nil, err
	}
	if len(tokens) != len(queries) {
		return nil, fmt.Errorf("tokenizer returned %d encodings for %d queries", len(tokens), len(queries))
	}
	return m.BatchEmbedQueriesFromTokens(tokens)
}

func (m *modelData) BatchEmbedQueriesFromTokens(queryTokens [][]int32) ([][]float32, error) {
	maxLength := 0
	for _, query := range queryTokens {
		if len(query) > maxLength {
			maxLength = len(query)
		}
	}
	if maxLength == 0 {
		return nil, fmt.Errorf("tokenizer returned empty encodings")
	}

	batchSize := len(queryTokens)
	shape := ort.NewShape(int64(batchSize), int64(maxLength))
	inputIds, attentionMask := batchTokenData(queryTokens, maxLength, int32(m.Tokenizer.Tokenizer.GetPadding().PadId))

	inputs := make([]ort.ArbitraryTensor, len(m.Input))
	defer destroyTensors(inputs)

	for i, inputInfo := range m.Input {
		tensor, err := buildBatchInputTensor(inputInfo, shape, inputIds, attentionMask)
		if err != nil {
			return nil, err
		}
		inputs[i] = tensor
	}

	outputs, err := m.run(inputs)
	if err != nil {
		return nil, err
	}
	defer destroyTensors(outputs)

	outputEmbeddings, err := extractBatchEmbeddings(outputs, batchSize, attentionMask)
	if err != nil {
		return nil, err
	}

	return outputEmbeddings, nil
}

func batchTokenData(tokens [][]int32, maxLength int, paddingToken int32) ([]int32, []int32) {
	inputIds := make([]int32, len(tokens)*maxLength)
	attentionMask := make([]int32, len(tokens)*maxLength)
	for batchIndex, tokenIds := range tokens {
		start := batchIndex * maxLength
		copy(inputIds[start:start+maxLength], tokenIds)
		for tokenIndex, tokenID := range tokenIds {
			if tokenID != paddingToken {
				attentionMask[start+tokenIndex] = 1
			}
		}
		for tokenIndex := len(tokenIds); tokenIndex < maxLength; tokenIndex++ {
			inputIds[start+tokenIndex] = paddingToken
		}
	}
	return inputIds, attentionMask
}

func buildBatchInputTensor(info ort.InputOutputInfo, shape ort.Shape, inputIds, attentionMask []int32) (ort.ArbitraryTensor, error) {
	data := inputIds
	name := strings.ToLower(info.Name)
	switch {
	case strings.Contains(name, "mask"):
		data = attentionMask
	case strings.Contains(name, "type"):
		data = make([]int32, len(inputIds))
	case strings.Contains(name, "position"):
		data = batchPositionIds(shape, len(inputIds))
	}

	switch info.DataType {
	case ort.TensorElementDataTypeInt64:
		return ort.NewTensor(shape, data)
	case ort.TensorElementDataTypeInt32:
		return ort.NewTensor(shape, data)
	default:
		return nil, fmt.Errorf("unsupported input tensor type %s for %s", info.DataType.String(), info.Name)
	}
}

func batchPositionIds(shape ort.Shape, size int) []int32 {
	positions := make([]int32, size)
	if len(shape) < 2 {
		return positions
	}

	sequenceLength := int(shape[1])
	for i := range positions {
		positions[i] = int32(i % sequenceLength)
	}
	return positions
}

func (m *modelData) run(inputs []ort.ArbitraryTensor) ([]ort.ArbitraryTensor, error) {
	session, err := ort.NewDynamicAdvancedSession(
		m.ModelPath,
		tensorNames(m.Input),
		tensorNames(m.Output),
		nil,
	)
	if err != nil {
		return nil, err
	}
	defer session.Destroy()

	inputs64 := make([]ort.ArbitraryTensor, len(inputs))
	for i, input := range inputs {
		switch tensor := input.(type) {
		case *ort.Tensor[int32]:
			inputs64[i], err = ort.NewTensor(tensor.GetShape(), convertToInt64(tensor.GetData()))
			if err != nil {
				return nil, err
			}
		default:
			inputs64[i] = input
		}
	}

	outputs := make([]ort.ArbitraryTensor, len(m.Output))
	if err := session.Run(inputs64, outputs); err != nil {
		return nil, err
	}
	return outputs, nil
}

func tensorNames(infos []ort.InputOutputInfo) []string {
	names := make([]string, len(infos))
	for i, info := range infos {
		names[i] = info.Name
	}
	return names
}

type ModelDimensions struct {
	input  int
	output int
}

func (m *modelData) VectorDimensions() (ModelDimensions, error) {
	outputDimension, err := fixedDimension(m.Output)
	if err != nil {
		return ModelDimensions{}, fmt.Errorf("model has no fixed output dimension: %w", err)
	}

	inputDimension, err := fixedDimension(m.Input)
	if err != nil {
		inputDimension = m.Tokenizer.MaxSequenceLength
	}
	if inputDimension <= 0 {
		return ModelDimensions{}, fmt.Errorf("model has no fixed input sequence length and tokenizer provides none")
	}

	return ModelDimensions{
		input:  inputDimension,
		output: outputDimension,
	}, nil
}

func fixedDimension(infos []ort.InputOutputInfo) (int, error) {
	if len(infos) == 0 {
		return 0, fmt.Errorf("no tensor infos")
	}

	dimensions := infos[0].Dimensions
	if len(dimensions) == 0 {
		return 0, fmt.Errorf("no dimensions")
	}

	for index := len(dimensions) - 1; index >= 0; index-- {
		if dimensions[index] > 1 {
			return int(dimensions[index]), nil
		}
	}

	return 0, fmt.Errorf("no fixed dimensions in %v", dimensions)
}

func destroyTensors(tensors []ort.ArbitraryTensor) {
	for _, tensor := range tensors {
		if tensor != nil {
			tensor.Destroy()
		}
	}
}

func extractBatchEmbeddings(outputs []ort.ArbitraryTensor, batchSize int, attentionMask []int32) ([][]float32, error) {
	data, shape, err := extractTensorData(outputs)
	if err != nil {
		return nil, err
	}
	return collapseBatchEmbedding(data, shape, batchSize, attentionMask)
}

func collapseBatchEmbedding(data []float32, shape ort.Shape, batchSize int, attentionMask []int32) ([][]float32, error) {
	if batchSize <= 0 || len(shape) == 0 || int(shape[0]) != batchSize {
		return nil, fmt.Errorf("invalid batch embedding shape: %v", shape)
	}

	if len(shape) == 2 {
		hiddenSize := int(shape[1])
		if hiddenSize <= 0 || len(data) != batchSize*hiddenSize {
			return nil, fmt.Errorf("invalid batch embedding shape: %v", shape)
		}

		embeddings := make([][]float32, batchSize)
		for batchIndex := range embeddings {
			start := batchIndex * hiddenSize
			embeddings[batchIndex] = append([]float32(nil), data[start:start+hiddenSize]...)
		}
		return embeddings, nil
	}

	if len(shape) == 3 {
		sequenceLength := int(shape[1])
		hiddenSize := int(shape[2])
		if sequenceLength <= 0 || hiddenSize <= 0 || len(data) != batchSize*sequenceLength*hiddenSize {
			return nil, fmt.Errorf("invalid batch embedding shape: %v", shape)
		}
		if len(attentionMask) != 0 && len(attentionMask) != batchSize*sequenceLength {
			return nil, fmt.Errorf("invalid attention mask length: got %d, want %d", len(attentionMask), batchSize*sequenceLength)
		}

		embeddings := make([][]float32, batchSize)
		for batchIndex := range embeddings {
			embeddings[batchIndex] = make([]float32, hiddenSize)
			validTokenCount := 0
			for tokenIndex := 0; tokenIndex < sequenceLength; tokenIndex++ {
				if len(attentionMask) != 0 && attentionMask[batchIndex*sequenceLength+tokenIndex] == 0 {
					continue
				}
				baseIndex := (batchIndex*sequenceLength + tokenIndex) * hiddenSize
				for hiddenIndex := 0; hiddenIndex < hiddenSize; hiddenIndex++ {
					embeddings[batchIndex][hiddenIndex] += data[baseIndex+hiddenIndex]
				}
				validTokenCount++
			}
			if validTokenCount == 0 {
				return nil, fmt.Errorf("batch item %d has no valid tokens", batchIndex)
			}
			for hiddenIndex := range embeddings[batchIndex] {
				embeddings[batchIndex][hiddenIndex] /= float32(validTokenCount)
			}
		}
		return embeddings, nil
	}

	return nil, fmt.Errorf("unsupported batch embedding shape: %v", shape)
}

func buildInputTensor(info ort.InputOutputInfo, tokenIds []int32) (ort.ArbitraryTensor, error) {
	name := strings.ToLower(info.Name)
	shape := tensorShape(info, tokenIds)
	tensorSize := int(shape.FlattenedSize())
	if tensorSize <= 0 {
		return nil, fmt.Errorf("invalid input tensor shape for %s: %v", info.Name, shape)
	}

	data := inputTensorData(name, tensorSize, tokenIds)
	switch info.DataType {
	case ort.TensorElementDataTypeInt64:
		return ort.NewTensor(shape, data)
	case ort.TensorElementDataTypeInt32:
		return ort.NewTensor(shape, data)
	default:
		return nil, fmt.Errorf("unsupported input tensor type %s for %s", info.DataType.String(), info.Name)
	}
}

func inputTensorData(name string, tensorSize int, tokenIds []int32) []int32 {
	data := make([]int32, tensorSize)
	tokenCount := tensorSize
	if tokenCount > len(tokenIds) {
		tokenCount = len(tokenIds)
	}

	switch {
	case strings.Contains(name, "mask"):
		for i := 0; i < tokenCount; i++ {
			data[i] = 1
		}
	case strings.Contains(name, "position"):
		for i := 0; i < tokenCount; i++ {
			data[i] = int32(i)
		}
	case !strings.Contains(name, "type"):
		copy(data, tokenIds[:tokenCount])
	}
	return data
}

func tensorShape(info ort.InputOutputInfo, tokenIds []int32) ort.Shape {
	shape := make(ort.Shape, len(info.Dimensions))
	for i, dimension := range info.Dimensions {
		switch {
		case dimension > 0:
			shape[i] = dimension
		case i == 0:
			shape[i] = 1
		default:
			shape[i] = int64(len(tokenIds))
		}
	}

	if len(shape) == 0 {
		return ort.NewShape(int64(len(tokenIds)))
	}

	if len(shape) == 1 {
		shape[0] = int64(len(tokenIds))
		return shape
	}

	if len(shape) >= 2 {
		shape[0] = 1
		shape[1] = int64(len(tokenIds))
	}

	return shape
}

func extractEmbedding(outputs []ort.ArbitraryTensor) ([]float32, error) {
	data, shape, err := extractTensorData(outputs)
	if err != nil {
		return nil, err
	}

	return collapseEmbedding(data, shape)
}

func extractTensorData(outputs []ort.ArbitraryTensor) ([]float32, ort.Shape, error) {
	for _, output := range outputs {
		if output == nil {
			continue
		}

		switch tensor := output.(type) {
		case *ort.Tensor[float32]:
			return tensor.GetData(), tensor.GetShape(), nil
		case *ort.Tensor[float64]:
			return convertToFloat32(tensor.GetData()), tensor.GetShape(), nil
		case *ort.Tensor[int64]:
			return convertToFloat32(tensor.GetData()), tensor.GetShape(), nil
		case *ort.Tensor[int32]:
			return convertToFloat32(tensor.GetData()), tensor.GetShape(), nil
		}
	}

	return nil, nil, fmt.Errorf("model did not return a tensor output")
}

func convertToFloat32[T float64 | int32 | int64](data []T) []float32 {
	converted := make([]float32, len(data))
	for i, value := range data {
		converted[i] = float32(value)
	}
	return converted
}

func convertToInt32(data []int64) []int32 {
	converted := make([]int32, len(data))
	for i, value := range data {
		converted[i] = int32(value)
	}
	return converted
}

func convertToInt64(data []int32) []int64 {
	converted := make([]int64, len(data))
	for i, value := range data {
		converted[i] = int64(value)
	}
	return converted
}

func collapseEmbedding(data []float32, shape ort.Shape) ([]float32, error) {
	if len(shape) <= 1 || shape[0] != 1 || len(shape) > 3 {
		return data, nil
	}

	embeddings, err := collapseBatchEmbedding(data, shape, 1, nil)
	if err != nil {
		return nil, err
	}
	return embeddings[0], nil
}
