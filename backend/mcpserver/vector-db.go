package mcpserver

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"sync"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3"
)

type VectorDB struct {
	*sql.DB
}

type vectorDataRow struct {
	Chunk_ID int32
	Filename string
	text     string
	Vector   []float32
	Distance float32
}

func (vdb *VectorDB) CreateVectorTable(vectorDimension int) error {
	if vectorDimension <= 0 {
		return fmt.Errorf("vector dimension must be positive")
	}

	createTableSQL := fmt.Sprintf(`

	CREATE TABLE IF NOT EXISTS documents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		text TEXT NOT NULL,
		filename TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE VIRTUAL TABLE vectors USING vec0(
		id INTEGER PRIMARY KEY REFERENCES documents(id),
		embedding float[%d]
	);

	-- cascades to delete vectors when its referencing document is deleted
	CREATE TRIGGER IF NOT EXISTS cascade_delete_vectors
	AFTER DELETE ON documents
	BEGIN
		DELETE FROM vectors WHERE id = OLD.id;
	END;

	
	CREATE TABLE IF NOT EXISTS metadata (
		key TEXT PRIMARY KEY NOT NULL,
		value TEXT
	)
	`, vectorDimension)
	_, err := vdb.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("error creating vectors table: %w", err)
	}

	return nil
}

var registerSQLiteVec sync.Once

// creates a fresh empty vector database
func NewVectorDB(dbPath string, vectorDimensions int) (*VectorDB, error) {
	if exists, err := common.FsExists(dbPath, false); err != nil {
		return nil, fmt.Errorf("error checking if database exists: %w", err)
	} else if exists {
		return (&VectorDB{}).LoadVectorDB(dbPath)
	}

	return (&VectorDB{}).CreateVectorDB(dbPath, vectorDimensions)
}

func (vdb *VectorDB) CreateVectorDB(dbPath string, vectorDimensions int) (*VectorDB, error) {
	if dbPath == "" {
		return nil, fmt.Errorf("database path is required")
	}
	if vectorDimensions <= 0 {
		return nil, fmt.Errorf("vector dimension must be positive")
	}

	// Remove existing database files if they exist
	for _, path := range []string{dbPath, dbPath + "-shm", dbPath + "-wal"} {
		if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("error removing existing vector database: %w", err)
		}
	}

	new_vdb, err := vdb.openVectorDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening vector database: %w", err)
	}

	if err := new_vdb.CreateVectorTable(vectorDimensions); err != nil {
		new_vdb.Close()
		return nil, fmt.Errorf("error creating vector table: %w", err)
	}

	return new_vdb, nil
}

// opens the existing vector database
func (vdb *VectorDB) LoadVectorDB(dbPath string) (*VectorDB, error) {
	if _, err := os.Stat(dbPath); err != nil {
		return nil, fmt.Errorf("error checking vector database: %w", err)
	}

	return vdb.openVectorDB(dbPath)
}

func (vdb *VectorDB) openVectorDB(dbPath string) (*VectorDB, error) {
	// Auto registers sqlite-vec with every SQLite connection opened by this process.
	registerSQLiteVec.Do(sqlite_vec.Auto)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return &VectorDB{DB: db}, nil
}

func (vdb *VectorDB) Close() error {
	return vdb.DB.Close()
}

func (vdb *VectorDB) InsertVector(data vectorDataRow) error {
	tx, err := vdb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // fallback on error, canceled by tx.Commit()

	if err := vdb.insertVector(tx, data); err != nil {
		return err
	}

	return tx.Commit()
}

func (vdb *VectorDB) insertVector(tx *sql.Tx, data vectorDataRow) error {
	normalizedVector, err := vdb.normalizeVector(data.Vector)
	if err != nil {
		return fmt.Errorf("normalize vector: %w", err)
	}
	vector, err := sqlite_vec.SerializeFloat32(normalizedVector)
	if err != nil {
		return fmt.Errorf("serialize vector: %w", err)
	}

	var lastInsertID int32
	err = tx.QueryRow(`
		INSERT INTO documents (text, filename)
		VALUES (?, ?)
		RETURNING id;`, data.text, data.Filename).Scan(&lastInsertID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO vectors (id, embedding) 
		VALUES (?, ?);`, lastInsertID, vector)
	if err != nil {
		return err
	}

	return nil
}

func (vdb *VectorDB) DeleteVector(chunkID int32) error {
	tx, err := vdb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // fallback on error, canceled by tx.Commit()

	_, err = tx.Exec(`DELETE FROM vectors WHERE id = ?;`, chunkID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM documents WHERE id = ?;`, chunkID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (vdb *VectorDB) HybridSearch(queryVector []float32, topK int) ([]vectorDataRow, error) {
	if topK <= 0 {
		return nil, fmt.Errorf("top K must be positive")
	}

	normalizedQuery, err := vdb.normalizeVector(queryVector)
	if err != nil {
		return nil, fmt.Errorf("normalize query vector: %w", err)
	}
	query, err := sqlite_vec.SerializeFloat32(normalizedQuery)
	if err != nil {
		return nil, fmt.Errorf("error serializing query vector: %w", err)
	}

	rows, err := vdb.Query(`SELECT d.id, d.text, v.distance FROM vectors v
		JOIN documents d ON v.id = d.id
		WHERE v.embedding MATCH ? AND k = ? ORDER BY v.distance ASC;`,
		query, topK)

	if err != nil {
		return nil, fmt.Errorf("search vectors: %w", err)
	}
	defer rows.Close()

	results := make([]vectorDataRow, 0, topK)
	for rows.Next() {
		var result vectorDataRow
		if err := rows.Scan(&result.Chunk_ID, &result.text, &result.Distance); err != nil {
			return nil, fmt.Errorf("scan vector search result: %w", err)
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vector search results: %w", err)
	}

	return results, nil
}

func (vdb *VectorDB) WriteToDisk() error {
	_, err := vdb.Exec("PRAGMA wal_checkpoint(FULL);")
	if err != nil {
		return fmt.Errorf("error checkpointing WAL: %w", err)
	}

	return nil
}

func (vdb *VectorDB) normalizeVector(vector []float32) ([]float32, error) {
	var sumOfSquares float32
	for _, value := range vector {
		sumOfSquares += float32(value) * float32(value)
	}
	if sumOfSquares == 0 {
		return nil, fmt.Errorf("vector must not be zero-length")
	}

	normalized := make([]float32, len(vector))
	length := math.Sqrt(float64(sumOfSquares))
	for index, value := range vector {
		normalized[index] = value / float32(length)
	}
	return normalized, nil
}
