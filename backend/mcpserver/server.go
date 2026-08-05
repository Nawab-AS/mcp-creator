package mcpserver

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Server struct {
	VectorDB
	modelData
	mcpServer       *mcp.Server
	httpServer      *http.Server
	EmbeddingMargin int
	name            string
	dimensions      ModelDimensions
	Paused          bool
}

type searchInput struct {
	Query string `json:"query" jsonschema:"The query to search for."`
	TopK  int    `json:"top_k,omitempty" jsonschema:"The maximum number of matching chunks to return."`
}

type searchResult struct {
	ChunkID  int32   `json:"chunk_id"`
	Filename string  `json:"filename"`
	Text     string  `json:"text"`
	Distance float32 `json:"distance"`
}

type getChunkInput struct {
	ChunkID int32 `json:"chunk_id" jsonschema:"The chunk ID returned by the search tool."`
}

type getChunkResult struct {
	ChunkID  int32  `json:"chunk_id"`
	Filename string `json:"filename"`
	Text     string `json:"text"`
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

	return &Server{ // unassigned fields default to zero-values
		VectorDB:        *vdb,
		modelData:       model,
		dimensions:      vectorDimensions,
		EmbeddingMargin: 5, // default margin
		name:            projectName,
	}, nil
}

func (s *Server) Close() error {
	if err := s.VectorDB.Close(); err != nil {
		return fmt.Errorf("error closing vector database: %w", err)
	}
	return nil
}

func ScrappableFile(filePath string) (bool, error) {
	if exists, err := common.FsExists(filePath, false); err != nil {
		return false, fmt.Errorf("error checking if file exists: %w", err)
	} else if !exists {
		return false, fmt.Errorf("file does not exist: %s", filePath)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".md" || ext == ".txt" {
		return true, nil
	}
	return false, nil
}

func textContent(filePath string) (string, error) {
	if scrappable, err := ScrappableFile(filePath); err != nil {
		return "", err
	} else if !scrappable {
		return "", fmt.Errorf("file is not scrappable: %s", filePath)
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

func (s *Server) fileModified(filePath string) (bool, int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false, 0, fmt.Errorf("error getting file info: %w", err)
	}
	modTime := fileInfo.ModTime().UnixNano()

	// get last modified time from database
	tx, err := s.VectorDB.Begin()
	if err != nil {
		return false, 0, fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback() // fallback on error, canceled by tx.Commit()

	metadataKey := fmt.Sprintf("last_mod_%s", common.Hash(filePath))
	var lastMod int64
	err = tx.QueryRow(`SELECT value FROM metadata WHERE key = ?;`, metadataKey).Scan(&lastMod)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, 0, fmt.Errorf("error reading file metadata: %w", err)
	}
	return lastMod < modTime, modTime, nil
}

const addFileProgressUpdateInterval = 1000 * time.Millisecond

func (s *Server) addFile(filePath string, onProgress func(float32)) error {
	if onProgress == nil {
		onProgress = func(float32) {}
	}

	var progress atomic.Int32
	setProgress := func(value float32) {
		if value < 0 {
			value = 0
		} else if value > 1 {
			value = 1
		}
		progress.Store(int32(value * 1000))
	}
	setProgress(0)

	stopProgress := make(chan struct{})
	progressDone := make(chan struct{})
	go func() {
		defer close(progressDone)
		ticker := time.NewTicker(addFileProgressUpdateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				onProgress(float32(progress.Load()) / 1000)
			case <-stopProgress:
				return
			}
		}
	}()
	defer func() {
		close(stopProgress)
		<-progressDone
	}()

	modified, modTime, err := s.fileModified(filePath)
	if err != nil {
		return fmt.Errorf("error checking if file is modified: %w", err)
	} else if !modified {
		setProgress(1)
		onProgress(1)
		return nil
	}

	// fmt.Printf("Embedding file `%s`\n", filePath)
	setProgress(0.05)
	text, err := textContent(filePath)
	if err != nil {
		return fmt.Errorf("error extracting text from file: %w", err)
	} else if len(text) == 0 {
		return fmt.Errorf("file contains no text")
	} else {
		text = "<START OF FILE>\n" + text + "\n<END OF FILE>"
	}

	// delete any existing vectors for this file
	tx, err := s.VectorDB.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	setProgress(0.15)
	if err := s.removeFile(tx, filePath); err != nil {
		return fmt.Errorf("error removing existing file entries: %w", err)
	}
	setProgress(0.25)
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
	setProgress(0.45)
	// fmt.Printf("Split text into %d chunks of up to %d tokens with %d-token overlap\n", len(rows), chunkSize, chunkSize-step)

	embeddings, err := s.modelData.BatchEmbedQueriesFromTokens(rows)
	if err != nil {
		return fmt.Errorf("error generating embeddings: %w", err)
	}
	if len(embeddings) != len(rows) {
		return fmt.Errorf("generated %d embeddings for %d chunks", len(embeddings), len(rows))
	}
	setProgress(0.70)
	// fmt.Printf("Generated embeddings for %d rows. %d dimensions\n", len(embeddings), len(embeddings[0]))

	for i, embedding := range embeddings {
		if len(embeddings) > 0 {
			setProgress(0.70 + (0.25 * float32(i+1) / float32(len(embeddings))))
		}
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
	metadataKey := fmt.Sprintf("last_mod_%s", common.Hash(filePath))
	if _, err := tx.Exec(`INSERT OR REPLACE INTO metadata (key, value) VALUES (?, ?);`, metadataKey, modTime); err != nil {
		return fmt.Errorf("error saving file metadata: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing document deletion: %w", err)
	}
	setProgress(1)
	onProgress(1)
	return nil
}

func (s *Server) removeFile(tx *sql.Tx, filePath string) error {
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

	// Delete Dangling Vectors
	res, err = tx.Exec(`DELETE FROM vectors
		WHERE id NOT IN (SELECT id FROM documents);`)
	if err != nil {
		return fmt.Errorf("error deleting dangling vectors: %w", err)
	}

	if _, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}
	return nil
}

func (s *Server) IndexDir(dirPath string) error {
	// pre-fetch list of useable files + size
	files := make(map[string]int64)
	totalSize := int64(0)

	err := filepath.WalkDir(dirPath, func(path string, file fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if file.IsDir() {
			return nil
		}
		if scrappable, err := ScrappableFile(path); err != nil {
			fmt.Printf("Error checking if file is scrappable: %v\n", err)
			return nil
		} else if !scrappable {
			return nil
		}
		if size, err := file.Info(); err == nil {
			files[path] = size.Size()
			totalSize += files[path]
		} else {
			fmt.Printf("Error getting file info for %s: %v\n", path, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking directory: %w", err)
	}
	// fmt.Printf("Total size of files to index: %d bytes\n", totalSize)

	// If there's nothing to do, emit completed and return
	if totalSize == 0 {
		common.EmitEvent("completed", "project-index", s.name)
		return nil
	}

	// Process files in deterministic order
	paths := make([]string, 0, len(files))
	for p := range files {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	overallProgress := float32(0)
	common.EmitEvent("progress", "project-index", s.name, overallProgress)

	var processed int64 = 0
	for _, filePath := range paths {
		size := files[filePath]
		initialProgress := float32(processed) / float32(totalSize)

		if err := s.addFile(filePath, func(progress float32) {
			overallProgress = initialProgress + (progress * float32(size) / float32(totalSize))
			common.EmitEvent("progress", "project-index", s.name, overallProgress)
		}); err != nil {
			common.EmitEvent("error", "project-index", s.name)
			return fmt.Errorf("error indexing file %s: %w", filePath, err)
		}

		// mark file as processed
		processed += size
		overallProgress = float32(processed) / float32(totalSize)
		common.EmitEvent("progress", "project-index", s.name, overallProgress)
	}

	common.EmitEvent("completed", "project-index", s.name)
	return nil
}

func (s *Server) HybridSearch(query string, topK int) ([]vectorDataRow, error) {
	embedding, err := s.modelData.EmbedQuery(query)
	if err != nil {
		return nil, fmt.Errorf("error embedding query: %w", err)
	}

	rawResults, err := s.VectorDB.HybridSearch(embedding, topK)
	if err != nil {
		return nil, fmt.Errorf("error performing hybrid search: %w", err)
	}

	// post-process
	results := make([]vectorDataRow, len(rawResults))
	for i, rawResult := range rawResults {
		results[i] = vectorDataRow{
			Chunk_ID: rawResult.Chunk_ID,
			Filename: rawResult.Filename,
			text:     rawResult.text,
			Vector:   []float32{},
			Distance: rawResult.Distance,
		}
	}

	return results, nil
}

func (s *Server) GetChunkText(chunkID int32) (*vectorDataRow, error) {
	return s.VectorDB.GetVector(chunkID)
}

func (s *Server) StartServer(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	projectsDir, projectName, err := common.ProjectDBPath(s.name)
	if err != nil {
		return fmt.Errorf("error getting project DB path: %w", err)
	}
	s.IndexDir(filepath.Join(projectsDir, projectName))

	s.mcpServer = mcp.NewServer(&mcp.Implementation{
		Name:  s.name,
		Title: s.name,
		Description: fmt.Sprintf(`
# '%s' MCP server
This is a data-retrieval MCP server (open-source at %s)

# Tools
## 'search'
Search indexed document chunks for a query, uses hybrid semaintic and keyword search.
Inputs: query (string), top_k (int, optional)
Outputs: list of relevant chunks (chunk_id, filename, text, distance)

## 'get-chunk'
Get the full text and source filename for an indexed chunk. useful when a retrivied chunk is cut off.
	Inputs: chunk_id (int)
	Outputs: chunk_id, filename, text
		`, s.name, "github.com/Nawab-as/mcp-creator"),
	}, nil)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "search",
		Description: "Search indexed document chunks for a query.",
	}, s.searchTool)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get-chunk",
		Description: "Get the full text and source filename for an indexed chunk.",
	}, s.getChunkTool)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mcp.NewSSEHandler(func(*http.Request) *mcp.Server { return s.mcpServer }, nil),
	}

	listener, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		s.httpServer = nil
		return fmt.Errorf("error starting MCP server: %w", err)
	}

	go func() {
		if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Printf("MCP server stopped unexpectedly: %v\n", err)
		}
	}()

	return nil
}

func (s *Server) StopServer() {
	if s.httpServer != nil {
		_ = s.httpServer.Close()
		s.httpServer = nil
	}
	if s.mcpServer != nil {
		s.mcpServer = nil
	}
}

func (s *Server) searchTool(_ context.Context, _ *mcp.CallToolRequest, input searchInput) (*mcp.CallToolResult, []searchResult, error) {
	if s.Paused {
		return nil, nil, fmt.Errorf("server is currently processing data, please try again later")
	}
	if strings.TrimSpace(input.Query) == "" {
		return nil, nil, fmt.Errorf("query must not be empty")
	}
	if input.TopK == 0 {
		input.TopK = 5
	}
	if input.TopK < 0 {
		return nil, nil, fmt.Errorf("top_k must be positive")
	}

	rows, err := s.HybridSearch(input.Query, input.TopK)
	if err != nil {
		return nil, nil, err
	}

	results := make([]searchResult, len(rows))
	for index, row := range rows {
		results[index] = searchResult{
			ChunkID:  row.Chunk_ID,
			Filename: row.Filename,
			Text:     row.text,
			Distance: row.Distance,
		}
	}
	return nil, results, nil
}

func (s *Server) getChunkTool(_ context.Context, _ *mcp.CallToolRequest, input getChunkInput) (*mcp.CallToolResult, *getChunkResult, error) {
	if s.Paused {
		return nil, nil, fmt.Errorf("server is currently processing data, please try again later")
	}
	if input.ChunkID < 1 {
		return nil, nil, fmt.Errorf("chunk_id must be positive")
	}

	row, err := s.GetChunkText(input.ChunkID)
	if err != nil {
		return nil, nil, err
	}
	return nil, &getChunkResult{
		ChunkID:  input.ChunkID,
		Filename: row.Filename,
		Text:     row.text,
	}, nil
}
