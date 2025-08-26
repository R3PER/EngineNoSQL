package service

import (
	"enginenosql/internal/engine"
	"fmt"
	"os"
	"path/filepath"
)

// DatabaseService provides database operations for the frontend
type DatabaseService struct {
	engine *engine.Engine
}

// DatabaseInfo represents database information for the frontend
type DatabaseInfo struct {
	Name        string   `json:"name"`
	Collections []string `json:"collections"`
}

// CollectionInfo represents collection information for the frontend
type CollectionInfo struct {
	Name          string   `json:"name"`
	DocumentCount int      `json:"document_count"`
	Indexes       []string `json:"indexes"`
}

// DocumentResponse represents a document response for the frontend
type DocumentResponse struct {
	ID        string                 `json:"id"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

// QueryRequest represents a query request from the frontend
type QueryRequest struct {
	Database   string      `json:"database"`
	Collection string      `json:"collection"`
	Field      string      `json:"field"`
	Value      interface{} `json:"value"`
}

// InsertRequest represents an insert request from the frontend
type InsertRequest struct {
	Database   string                 `json:"database"`
	Collection string                 `json:"collection"`
	ID         string                 `json:"id"`
	Data       map[string]interface{} `json:"data"`
}

// UpdateRequest represents an update request from the frontend
type UpdateRequest struct {
	Database   string                 `json:"database"`
	Collection string                 `json:"collection"`
	ID         string                 `json:"id"`
	Data       map[string]interface{} `json:"data"`
}

// DeleteRequest represents a delete request from the frontend
type DeleteRequest struct {
	Database   string `json:"database"`
	Collection string `json:"collection"`
	ID         string `json:"id"`
}

// NewDatabaseService creates a new database service for a specific user
func NewDatabaseService(userID string) *DatabaseService {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic("Failed to get home directory: " + err.Error())
	}

	// Create user-specific data directory
	dataDir := filepath.Join(homeDir, ".enginenosql", "data", userID)
	engine := engine.NewEngine(dataDir)

	return &DatabaseService{
		engine: engine,
	}
}

// CreateDatabase creates a new database
func (s *DatabaseService) CreateDatabase(name string) error {
	if name == "" {
		return fmt.Errorf("database name cannot be empty")
	}
	return s.engine.CreateDatabase(name)
}

// DeleteDatabase deletes a database
func (s *DatabaseService) DeleteDatabase(name string) error {
	if name == "" {
		return fmt.Errorf("database name cannot be empty")
	}
	return s.engine.DeleteDatabase(name)
}

// ListDatabases returns list of all databases
func (s *DatabaseService) ListDatabases() ([]DatabaseInfo, error) {
	names := s.engine.ListDatabases()
	var databases []DatabaseInfo

	for _, name := range names {
		db, err := s.engine.GetDatabase(name)
		if err != nil {
			continue // Skip databases that can't be loaded
		}

		var collections []string
		for collName := range db.Collections {
			collections = append(collections, collName)
		}

		databases = append(databases, DatabaseInfo{
			Name:        name,
			Collections: collections,
		})
	}

	return databases, nil
}

// CreateCollection creates a new collection in a database
func (s *DatabaseService) CreateCollection(dbName, collName string) error {
	if dbName == "" || collName == "" {
		return fmt.Errorf("database and collection names cannot be empty")
	}

	db, err := s.engine.GetDatabase(dbName)
	if err != nil {
		return err
	}

	err = db.CreateCollection(collName)
	if err != nil {
		return err
	}

	return s.engine.SaveDatabase(dbName)
}

// DeleteCollection deletes a collection from a database
func (s *DatabaseService) DeleteCollection(dbName, collName string) error {
	if dbName == "" || collName == "" {
		return fmt.Errorf("database and collection names cannot be empty")
	}

	db, err := s.engine.GetDatabase(dbName)
	if err != nil {
		return err
	}

	// Remove collection from database
	delete(db.Collections, collName)

	return s.engine.SaveDatabase(dbName)
}

// GetCollections returns list of collections in a database
func (s *DatabaseService) GetCollections(dbName string) ([]CollectionInfo, error) {
	if dbName == "" {
		return nil, fmt.Errorf("database name cannot be empty")
	}

	db, err := s.engine.GetDatabase(dbName)
	if err != nil {
		return nil, err
	}

	var collections []CollectionInfo
	for name, collection := range db.Collections {
		var indexes []string
		for field := range collection.Indexes {
			indexes = append(indexes, field)
		}

		collections = append(collections, CollectionInfo{
			Name:          name,
			DocumentCount: len(collection.Documents),
			Indexes:       indexes,
		})
	}

	return collections, nil
}

// InsertDocument inserts a document into a collection
func (s *DatabaseService) InsertDocument(req InsertRequest) error {
	if req.Database == "" || req.Collection == "" || req.ID == "" {
		return fmt.Errorf("database, collection, and document ID cannot be empty")
	}

	db, err := s.engine.GetDatabase(req.Database)
	if err != nil {
		return err
	}

	collection, err := db.GetCollection(req.Collection)
	if err != nil {
		return err
	}

	err = collection.Insert(req.ID, req.Data)
	if err != nil {
		return err
	}

	return s.engine.SaveDatabase(req.Database)
}

// UpdateDocument updates a document in a collection
func (s *DatabaseService) UpdateDocument(req UpdateRequest) error {
	if req.Database == "" || req.Collection == "" || req.ID == "" {
		return fmt.Errorf("database, collection, and document ID cannot be empty")
	}

	db, err := s.engine.GetDatabase(req.Database)
	if err != nil {
		return err
	}

	collection, err := db.GetCollection(req.Collection)
	if err != nil {
		return err
	}

	err = collection.Update(req.ID, req.Data)
	if err != nil {
		return err
	}

	return s.engine.SaveDatabase(req.Database)
}

// DeleteDocument deletes a document from a collection
func (s *DatabaseService) DeleteDocument(req DeleteRequest) error {
	if req.Database == "" || req.Collection == "" || req.ID == "" {
		return fmt.Errorf("database, collection, and document ID cannot be empty")
	}

	db, err := s.engine.GetDatabase(req.Database)
	if err != nil {
		return err
	}

	collection, err := db.GetCollection(req.Collection)
	if err != nil {
		return err
	}

	err = collection.Delete(req.ID)
	if err != nil {
		return err
	}

	return s.engine.SaveDatabase(req.Database)
}

// QueryDocuments queries documents in a collection
func (s *DatabaseService) QueryDocuments(req QueryRequest) ([]DocumentResponse, error) {
	if req.Database == "" || req.Collection == "" {
		return nil, fmt.Errorf("database and collection names cannot be empty")
	}

	db, err := s.engine.GetDatabase(req.Database)
	if err != nil {
		return nil, err
	}

	collection, err := db.GetCollection(req.Collection)
	if err != nil {
		return nil, err
	}

	var documents []*engine.Document
	if req.Field == "" {
		// Get all documents if no field specified
		documents = collection.GetAll()
	} else {
		// Query by field and value
		documents, err = collection.Find(req.Field, req.Value)
		if err != nil {
			return nil, err
		}
	}

	var response []DocumentResponse
	for _, doc := range documents {
		response = append(response, DocumentResponse{
			ID:        doc.ID,
			Data:      doc.Data,
			CreatedAt: doc.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: doc.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return response, nil
}

// CreateIndex creates an index on a field in a collection
func (s *DatabaseService) CreateIndex(dbName, collName, field string) error {
	if dbName == "" || collName == "" || field == "" {
		return fmt.Errorf("database, collection, and field names cannot be empty")
	}

	db, err := s.engine.GetDatabase(dbName)
	if err != nil {
		return err
	}

	collection, err := db.GetCollection(collName)
	if err != nil {
		return err
	}

	collection.CreateIndex(field)
	return s.engine.SaveDatabase(dbName)
}

// GetDatabaseStats returns statistics about a database
func (s *DatabaseService) GetDatabaseStats(dbName string) (map[string]interface{}, error) {
	if dbName == "" {
		return nil, fmt.Errorf("database name cannot be empty")
	}

	db, err := s.engine.GetDatabase(dbName)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]interface{})
	stats["name"] = db.Name
	stats["collections_count"] = len(db.Collections)

	totalDocuments := 0
	totalIndexes := 0
	for _, collection := range db.Collections {
		totalDocuments += len(collection.Documents)
		totalIndexes += len(collection.Indexes)
	}

	stats["total_documents"] = totalDocuments
	stats["total_indexes"] = totalIndexes

	return stats, nil
}

// Advanced Query Support

// AdvancedQueryRequest represents an advanced query request
type AdvancedQueryRequest struct {
	Database   string        `json:"database"`
	Collection string        `json:"collection"`
	Filters    []QueryFilter `json:"filters"`
	Sort       *SortOption   `json:"sort"`
	Limit      int           `json:"limit"`
	Skip       int           `json:"skip"`
}

// QueryFilter represents a query filter
type QueryFilter struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// SortOption represents sorting options
type SortOption struct {
	Field     string `json:"field"`
	Ascending bool   `json:"ascending"`
}

// AdvancedQuery executes an advanced query with filters, sorting, and pagination
func (s *DatabaseService) AdvancedQuery(req AdvancedQueryRequest) ([]DocumentResponse, error) {
	if req.Database == "" || req.Collection == "" {
		return nil, fmt.Errorf("database and collection names cannot be empty")
	}

	db, err := s.engine.GetDatabase(req.Database)
	if err != nil {
		return nil, err
	}

	collection, err := db.GetCollection(req.Collection)
	if err != nil {
		return nil, err
	}

	// Build query
	query := collection.NewQuery()

	// Add filters
	for _, filter := range req.Filters {
		query = query.Where(filter.Field, filter.Operator, filter.Value)
	}

	// Add sorting
	if req.Sort != nil {
		query = query.Sort(req.Sort.Field, req.Sort.Ascending)
	}

	// Add pagination
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}
	if req.Skip > 0 {
		query = query.Skip(req.Skip)
	}

	// Execute query
	documents, err := query.Execute()
	if err != nil {
		return nil, err
	}

	var response []DocumentResponse
	for _, doc := range documents {
		response = append(response, DocumentResponse{
			ID:        doc.ID,
			Data:      doc.Data,
			CreatedAt: doc.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: doc.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return response, nil
}

// CountDocuments counts documents matching the query
func (s *DatabaseService) CountDocuments(req AdvancedQueryRequest) (int, error) {
	if req.Database == "" || req.Collection == "" {
		return 0, fmt.Errorf("database and collection names cannot be empty")
	}

	db, err := s.engine.GetDatabase(req.Database)
	if err != nil {
		return 0, err
	}

	collection, err := db.GetCollection(req.Collection)
	if err != nil {
		return 0, err
	}

	// Build query
	query := collection.NewQuery()

	// Add filters
	for _, filter := range req.Filters {
		query = query.Where(filter.Field, filter.Operator, filter.Value)
	}

	return query.Count()
}

// Import/Export Support

// ExportRequest represents an export request
type ExportRequest struct {
	Database   string                `json:"database"`
	Collection string                `json:"collection"`
	Format     string                `json:"format"`
	Query      *AdvancedQueryRequest `json:"query"`
	FilePath   string                `json:"file_path"`
}

// ImportRequest represents an import request
type ImportRequest struct {
	Database         string `json:"database"`
	Collection       string `json:"collection"`
	Format           string `json:"format"`
	FilePath         string `json:"file_path"`
	CreateCollection bool   `json:"create_collection"`
	OverwriteData    bool   `json:"overwrite_data"`
	IDField          string `json:"id_field"`
}

// ExportData exports data from a collection
func (s *DatabaseService) ExportData(req ExportRequest) error {
	importExportManager := engine.NewImportExportManager(s.engine)

	options := engine.ExportOptions{
		Format:     engine.ImportFormat(req.Format),
		Collection: req.Collection,
		FilePath:   req.FilePath,
	}

	// Add query if specified
	if req.Query != nil {
		db, err := s.engine.GetDatabase(req.Database)
		if err != nil {
			return err
		}

		collection, err := db.GetCollection(req.Collection)
		if err != nil {
			return err
		}

		query := collection.NewQuery()
		for _, filter := range req.Query.Filters {
			query = query.Where(filter.Field, filter.Operator, filter.Value)
		}

		if req.Query.Sort != nil {
			query = query.Sort(req.Query.Sort.Field, req.Query.Sort.Ascending)
		}

		if req.Query.Limit > 0 {
			query = query.Limit(req.Query.Limit)
		}

		if req.Query.Skip > 0 {
			query = query.Skip(req.Query.Skip)
		}

		options.Query = query
	}

	return importExportManager.ExportData(req.Database, options)
}

// ImportData imports data into a collection
func (s *DatabaseService) ImportData(req ImportRequest) (*engine.ImportResult, error) {
	importExportManager := engine.NewImportExportManager(s.engine)

	options := engine.ImportOptions{
		Format:           engine.ImportFormat(req.Format),
		Collection:       req.Collection,
		FilePath:         req.FilePath,
		CreateCollection: req.CreateCollection,
		OverwriteData:    req.OverwriteData,
		IDField:          req.IDField,
	}

	result, err := importExportManager.ImportData(req.Database, options)
	if err != nil {
		return nil, err
	}

	// Save database after import
	if err := s.engine.SaveDatabase(req.Database); err != nil {
		return nil, fmt.Errorf("failed to save database after import: %v", err)
	}

	return result, nil
}

// ImportDataFromContent imports data from file content
func (s *DatabaseService) ImportDataFromContent(database, collection, content, format string, createCollection bool) (*engine.ImportResult, error) {
	importExportManager := engine.NewImportExportManager(s.engine)

	result, err := importExportManager.ImportDataFromContent(database, collection, content, format, createCollection)
	if err != nil {
		return nil, err
	}

	// Save database after import
	if err := s.engine.SaveDatabase(database); err != nil {
		return nil, fmt.Errorf("failed to save database after import: %v", err)
	}

	return result, nil
}

// GetSupportedFormats returns supported import/export formats
func (s *DatabaseService) GetSupportedFormats() []string {
	importExportManager := engine.NewImportExportManager(s.engine)
	return importExportManager.GetSupportedFormats()
}

// Backup Support

// BackupRequest represents a backup request
type BackupRequest struct {
	Database   string `json:"database"`
	BackupName string `json:"backup_name"`
}

// RestoreRequest represents a restore request
type RestoreRequest struct {
	BackupPath string `json:"backup_path"`
	NewDbName  string `json:"new_db_name"`
}

// CreateBackup creates a backup of a database
func (s *DatabaseService) CreateBackup(req BackupRequest) (*engine.BackupInfo, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	backupDir := filepath.Join(homeDir, ".enginenosql", "backups")
	backupManager := engine.NewBackupManager(s.engine, backupDir)

	return backupManager.CreateBackup(req.Database, req.BackupName)
}

// RestoreBackup restores a database from backup
func (s *DatabaseService) RestoreBackup(req RestoreRequest) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}

	backupDir := filepath.Join(homeDir, ".enginenosql", "backups")
	backupManager := engine.NewBackupManager(s.engine, backupDir)

	return backupManager.RestoreBackup(req.BackupPath, req.NewDbName)
}

// ListBackups lists all available backups
func (s *DatabaseService) ListBackups() ([]engine.BackupInfo, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	backupDir := filepath.Join(homeDir, ".enginenosql", "backups")
	backupManager := engine.NewBackupManager(s.engine, backupDir)

	return backupManager.ListBackups()
}

// CompactDatabase compacts a database to optimize storage
func (s *DatabaseService) CompactDatabase(dbName string) error {
	if dbName == "" {
		return fmt.Errorf("database name cannot be empty")
	}

	return s.engine.CompactDatabase(dbName)
}

// GetDetailedDatabaseStats returns detailed statistics about a database
func (s *DatabaseService) GetDetailedDatabaseStats(dbName string) (*engine.DatabaseStats, error) {
	if dbName == "" {
		return nil, fmt.Errorf("database name cannot be empty")
	}

	return s.engine.GetDatabaseStats(dbName)
}
