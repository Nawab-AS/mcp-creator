package mcpserver

import (
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type Server struct {
	VectorDB
	modelData
	EmbeddingMargin int
	dimensions      ModelDimensions
}

func NewServer(projectName string, modelName string) (*Server, error) {
	projectDir, projectDB, err := common.ProjectDBPath(projectName)
	if err != nil {
		return nil, err
	}

	// load model
	model, err := loadModel(modelName)
	if err != nil {
		return nil, fmt.Errorf("error loading model: %w", err)
	}

	// instantiate new vector database
	vectorDimensions, err := model.VectorDimensions()
	if err != nil {
		return nil, fmt.Errorf("error getting vector dimension: %w", err)
	}
	vdb, err := NewVectorDB(filepath.Join(projectDir, projectDB), vectorDimensions.output)
	if err != nil {
		return nil, fmt.Errorf("error creating vector database: %w", err)
	}

	return &Server{
		VectorDB:        *vdb,
		modelData:       model,
		dimensions:      vectorDimensions,
		EmbeddingMargin: 5, // default margin
	}, nil
}

func (s *Server) Close() error {
	if err := s.VectorDB.Close(); err != nil {
		return fmt.Errorf("error closing vector database: %w", err)
	}
	return nil
}

func textContent(filePath string) (string, error) {
	if exists, err := common.FsExists(filePath, false); err != nil {
		return "", fmt.Errorf("error checking if file exists: %w", err)
	} else if !exists {
		return "", fmt.Errorf("file does not exist: %s", filePath)
	}

	switch filepath.Ext(filePath) {
	case ".md": // LLMs natively support markdown, just read is text
		fallthrough
	case ".txt": //text file
		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("error reading text file: %w", err)
		}
		return string(content), nil
	// TODO: add support for more file types
	default:
		return "", fmt.Errorf("unsupported file type: %s", filepath.Ext(filePath))
	}
}

func (s *Server) AddFile(filePath string) error {
	// fmt.Printf("Embedding file `%s`\n", filePath)
	text, err := textContent(filePath)
	if err != nil {
		return fmt.Errorf("error extracting text from file: %w", err)
	} else if len(text) == 0 {
		return fmt.Errorf("file contains no text")
	} else {
		text = "<START OF FILE>\n" + text + "\n<END OF FILE>"
	}

	s.RemoveFile(filePath)
	tokens, err := s.modelData.Tokenize(text)
	if err != nil {
		return fmt.Errorf("error tokenizing text: %w", err)
	}
	// fmt.Printf("Tokenized text into %d tokens\n", len(tokens))

	chunkSize := s.dimensions.input
	if chunkSize <= 0 {
		chunkSize = len(tokens)
	}
	if chunkSize <= 0 {
		return fmt.Errorf("unable to determine a valid chunk size")
	}

	var rows [][]int32
	var chunkText []string

	step := chunkSize - s.EmbeddingMargin*2
	if step < 1 {
		step = 1
	}

	for start := 0; start < len(tokens); start += step {
		end := min(start+chunkSize, len(tokens))
		chunkTokens := tokens[start:end]
		row := make([]int32, chunkSize)
		copy(row, chunkTokens)
		rows = append(rows, row)
		chunkText = append(chunkText, s.modelData.Tokenizer.Decode(chunkTokens))
	}
	// fmt.Printf("Split text into %d chunks of up to %d tokens with %d-token overlap\n", len(rows), chunkSize, chunkSize-step)

	embeddings, err := s.modelData.BatchEmbedQueriesFromTokens(rows)
	if err != nil {
		return fmt.Errorf("error generating embeddings: %w", err)
	}
	if len(embeddings) != len(rows) {
		return fmt.Errorf("generated %d embeddings for %d chunks", len(embeddings), len(rows))
	}
	// fmt.Printf("Generated embeddings for %d rows. %d dimensions\n", len(embeddings), len(embeddings[0]))

	tx, err := s.VectorDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // fallback on error, canceled by tx.Commit()

	for i, embedding := range embeddings {
		if err := s.VectorDB.insertVector(tx, vectorDataRow{
			Chunk_ID: int32(i),
			Filename: filePath,
			text:     chunkText[i],
			Vector:   embedding,
		}); err != nil {
			return fmt.Errorf("error inserting vector for row %d: %w", i, err)
		}
	}
	// fmt.Printf("Inserted %d vectors into the database\n", len(embeddings))

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing document deletion: %w", err)
	}
	return nil
}

func (s *Server) RemoveFile(filePath string) error {
	tx, err := s.VectorDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // fallback on error, canceled by tx.Commit()

	// Delete vectors
	res, err := tx.Exec(`
		DELETE FROM vectors
		WHERE id IN (SELECT id FROM documents WHERE filename = ?);`, filePath)
	if err != nil {
		return fmt.Errorf("error deleting vectors: %w", err)
	}
	if _, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("error checking vectors rows affected: %w", err)
	}

	// Delete documents
	res, err = tx.Exec(`DELETE FROM documents WHERE filename = ?;`, filePath)
	if err != nil {
		return fmt.Errorf("error deleting existing document entries: %w", err)
	}

	if _, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing document deletion: %w", err)
	}
	return nil
}

func (s *Server) HybridSearch(query string, topK int) ([]vectorDataRow, error) {
	embedding, err := s.modelData.EmbedQuery(query)
	if err != nil {
		return nil, fmt.Errorf("error embedding query: %w", err)
	}

	results, err := s.VectorDB.HybridSearch(embedding, topK)
	if err != nil {
		return nil, fmt.Errorf("error performing hybrid search: %w", err)
	}
	return results, nil
}
