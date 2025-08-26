package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Document represents a document in the database
type Document struct {
	ID        string                 `json:"_id"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// Collection represents a collection of documents
type Collection struct {
	Name      string               `json:"name"`
	Documents map[string]*Document `json:"documents"`
	Indexes   map[string]Index     `json:"indexes"`
	mutex     sync.RWMutex
}

// Index represents an index on a field
type Index struct {
	Field  string            `json:"field"`
	Values map[string]string `json:"values"` // value -> document_id
}

// Database represents the main database structure
type Database struct {
	Name        string                 `json:"name"`
	Collections map[string]*Collection `json:"collections"`
	Path        string                 `json:"path"`
	mutex       sync.RWMutex
}

// Engine represents the NoSQL database engine
type Engine struct {
	databases map[string]*Database
	dataDir   string
	mutex     sync.RWMutex
}

// NewEngine creates a new database engine
func NewEngine(dataDir string) *Engine {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create data directory: %v", err))
	}

	return &Engine{
		databases: make(map[string]*Database),
		dataDir:   dataDir,
	}
}

// CreateDatabase creates a new database
func (e *Engine) CreateDatabase(name string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if _, exists := e.databases[name]; exists {
		return fmt.Errorf("database '%s' already exists", name)
	}

	dbPath := filepath.Join(e.dataDir, name+".enosql")
	db := &Database{
		Name:        name,
		Collections: make(map[string]*Collection),
		Path:        dbPath,
	}

	e.databases[name] = db
	return e.saveDatabase(db)
}

// GetDatabase retrieves a database by name
func (e *Engine) GetDatabase(name string) (*Database, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	if db, exists := e.databases[name]; exists {
		return db, nil
	}

	// Try to load from file
	return e.loadDatabase(name)
}

// CreateCollection creates a new collection in a database
func (db *Database) CreateCollection(name string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, exists := db.Collections[name]; exists {
		return fmt.Errorf("collection '%s' already exists", name)
	}

	db.Collections[name] = &Collection{
		Name:      name,
		Documents: make(map[string]*Document),
		Indexes:   make(map[string]Index),
	}

	return nil
}

// GetCollection retrieves a collection by name
func (db *Database) GetCollection(name string) (*Collection, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	if collection, exists := db.Collections[name]; exists {
		return collection, nil
	}

	return nil, fmt.Errorf("collection '%s' not found", name)
}

// Insert inserts a document into the collection
func (c *Collection) Insert(id string, data map[string]interface{}) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, exists := c.Documents[id]; exists {
		return fmt.Errorf("document with id '%s' already exists", id)
	}

	doc := &Document{
		ID:        id,
		Data:      data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	c.Documents[id] = doc
	c.updateIndexes(doc)

	return nil
}

// Update updates a document in the collection
func (c *Collection) Update(id string, data map[string]interface{}) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	doc, exists := c.Documents[id]
	if !exists {
		return fmt.Errorf("document with id '%s' not found", id)
	}

	// Remove old indexes
	c.removeFromIndexes(doc)

	// Update document
	doc.Data = data
	doc.UpdatedAt = time.Now()

	// Update indexes
	c.updateIndexes(doc)

	return nil
}

// Delete deletes a document from the collection
func (c *Collection) Delete(id string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	doc, exists := c.Documents[id]
	if !exists {
		return fmt.Errorf("document with id '%s' not found", id)
	}

	c.removeFromIndexes(doc)
	delete(c.Documents, id)

	return nil
}

// Find finds documents by field value
func (c *Collection) Find(field string, value interface{}) ([]*Document, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var results []*Document

	// Check if there's an index for this field
	if index, exists := c.Indexes[field]; exists {
		valueStr := fmt.Sprintf("%v", value)
		if docID, found := index.Values[valueStr]; found {
			if doc, exists := c.Documents[docID]; exists {
				results = append(results, doc)
			}
		}
	} else {
		// Full scan if no index
		for _, doc := range c.Documents {
			if docValue, exists := doc.Data[field]; exists {
				if fmt.Sprintf("%v", docValue) == fmt.Sprintf("%v", value) {
					results = append(results, doc)
				}
			}
		}
	}

	return results, nil
}

// GetAll returns all documents in the collection
func (c *Collection) GetAll() []*Document {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var docs []*Document
	for _, doc := range c.Documents {
		docs = append(docs, doc)
	}

	return docs
}

// CreateIndex creates an index on a field
func (c *Collection) CreateIndex(field string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	index := Index{
		Field:  field,
		Values: make(map[string]string),
	}

	// Build index from existing documents
	for _, doc := range c.Documents {
		if value, exists := doc.Data[field]; exists {
			valueStr := fmt.Sprintf("%v", value)
			index.Values[valueStr] = doc.ID
		}
	}

	c.Indexes[field] = index
}

// updateIndexes updates indexes when a document is inserted/updated
func (c *Collection) updateIndexes(doc *Document) {
	for field, index := range c.Indexes {
		if value, exists := doc.Data[field]; exists {
			valueStr := fmt.Sprintf("%v", value)
			index.Values[valueStr] = doc.ID
			c.Indexes[field] = index
		}
	}
}

// removeFromIndexes removes document from indexes when deleted
func (c *Collection) removeFromIndexes(doc *Document) {
	for field, index := range c.Indexes {
		if value, exists := doc.Data[field]; exists {
			valueStr := fmt.Sprintf("%v", value)
			delete(index.Values, valueStr)
			c.Indexes[field] = index
		}
	}
}

// saveDatabase saves database to file with .enosql extension
func (e *Engine) saveDatabase(db *Database) error {
	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal database: %v", err)
	}

	return os.WriteFile(db.Path, data, 0644)
}

// loadDatabase loads database from file
func (e *Engine) loadDatabase(name string) (*Database, error) {
	dbPath := filepath.Join(e.dataDir, name+".enosql")

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("database '%s' not found", name)
	}

	data, err := os.ReadFile(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read database file: %v", err)
	}

	var db Database
	if err := json.Unmarshal(data, &db); err != nil {
		return nil, fmt.Errorf("failed to unmarshal database: %v", err)
	}

	// Initialize mutexes for collections
	for _, collection := range db.Collections {
		collection.mutex = sync.RWMutex{}
	}

	e.databases[name] = &db
	return &db, nil
}

// SaveDatabase saves a database to disk
func (e *Engine) SaveDatabase(name string) error {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	db, exists := e.databases[name]
	if !exists {
		return fmt.Errorf("database '%s' not found", name)
	}

	return e.saveDatabase(db)
}

// DeleteDatabase deletes a database and its file
func (e *Engine) DeleteDatabase(name string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Remove from memory
	delete(e.databases, name)

	// Remove file
	dbPath := filepath.Join(e.dataDir, name+".enosql")
	if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete database file: %v", err)
	}

	return nil
}

// ListDatabases returns list of all databases
func (e *Engine) ListDatabases() []string {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	var names []string
	for name := range e.databases {
		names = append(names, name)
	}

	// Also check for .enosql files in data directory
	files, err := filepath.Glob(filepath.Join(e.dataDir, "*.enosql"))
	if err == nil {
		for _, file := range files {
			name := filepath.Base(file)
			name = name[:len(name)-7] // Remove .enosql extension
			found := false
			for _, existing := range names {
				if existing == name {
					found = true
					break
				}
			}
			if !found {
				names = append(names, name)
			}
		}
	}

	return names
}
