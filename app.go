package main

import (
	"context"
	"enginenosql/internal/auth"
	"enginenosql/internal/engine"
	"enginenosql/internal/service"
)

// App struct
type App struct {
	ctx         context.Context
	authService *auth.AuthService
	dbServices  map[string]*service.DatabaseService // sessionID -> service
}

// NewApp creates a new App application struct
func NewApp() *App {
	authService, err := auth.NewAuthService()
	if err != nil {
		panic("Failed to initialize auth service: " + err.Error())
	}

	return &App{
		authService: authService,
		dbServices:  make(map[string]*service.DatabaseService),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Authentication operations

// Register registers a new user
func (a *App) Register(req auth.RegisterRequest) (*auth.LoginResponse, error) {
	return a.authService.Register(req)
}

// Login authenticates a user and creates a session
func (a *App) Login(req auth.LoginRequest) (*auth.LoginResponse, error) {
	response, err := a.authService.Login(req)
	if err != nil {
		return nil, err
	}

	// Create database service for the user
	if response.Success && response.SessionID != "" && response.User != nil {
		a.dbServices[response.User.ID] = service.NewDatabaseService(response.User.ID)
	}

	return response, nil
}

// Logout logs out a user
func (a *App) Logout(sessionID string) error {
	// Remove database service
	delete(a.dbServices, sessionID)

	// Logout from auth service
	return a.authService.Logout(sessionID)
}

// ValidateSession validates a session
func (a *App) ValidateSession(sessionID string) (*auth.Session, error) {
	return a.authService.ValidateSession(sessionID)
}

// getDBService gets the database service for a session
func (a *App) getDBService(sessionID string) (*service.DatabaseService, error) {
	// Validate session and get user info
	session, err := a.authService.ValidateSession(sessionID)
	if err != nil {
		return nil, err
	}

	// Get or create database service for user
	if dbService, exists := a.dbServices[session.UserID]; exists {
		return dbService, nil
	}

	// Create new database service for user
	dbService := service.NewDatabaseService(session.UserID)
	a.dbServices[session.UserID] = dbService
	return dbService, nil
}

// Database operations exposed to frontend (require session)

// CreateDatabase creates a new database
func (a *App) CreateDatabase(sessionID, name string) error {
	dbService, err := a.getDBService(sessionID)
	if err != nil {
		return err
	}
	return dbService.CreateDatabase(name)
}

// DeleteDatabase deletes a database
func (a *App) DeleteDatabase(sessionID, name string) error {
	dbService, err := a.getDBService(sessionID)
	if err != nil {
		return err
	}
	return dbService.DeleteDatabase(name)
}

// ListDatabases returns list of all databases
func (a *App) ListDatabases(sessionID string) ([]service.DatabaseInfo, error) {
	dbService, err := a.getDBService(sessionID)
	if err != nil {
		return nil, err
	}
	return dbService.ListDatabases()
}

// CreateCollection creates a new collection in a database
func (a *App) CreateCollection(sessionID, dbName, collName string) error {
	dbService, err := a.getDBService(sessionID)
	if err != nil {
		return err
	}
	return dbService.CreateCollection(dbName, collName)
}

// DeleteCollection deletes a collection from a database
func (a *App) DeleteCollection(sessionID, dbName, collName string) error {
	dbService, err := a.getDBService(sessionID)
	if err != nil {
		return err
	}
	return dbService.DeleteCollection(dbName, collName)
}

// GetCollections returns list of collections in a database
func (a *App) GetCollections(sessionID, dbName string) ([]service.CollectionInfo, error) {
	dbService, err := a.getDBService(sessionID)
	if err != nil {
		return nil, err
	}
	return dbService.GetCollections(dbName)
}

// InsertDocument inserts a document into a collection
func (a *App) InsertDocument(sessionID string, req service.InsertRequest) error {
	dbService, err := a.getDBService(sessionID)
	if err != nil {
		return err
	}
	return dbService.InsertDocument(req)
}

// UpdateDocument updates a document in a collection
func (a *App) UpdateDocument(sessionID string, req service.UpdateRequest) error {
	dbService, err := a.getDBService(sessionID)
	if err != nil {
		return err
	}
	return dbService.UpdateDocument(req)
}

// DeleteDocument deletes a document from a collection
func (a *App) DeleteDocument(sessionID string, req service.DeleteRequest) error {
	dbService, err := a.getDBService(sessionID)
	if err != nil {
		return err
	}
	return dbService.DeleteDocument(req)
}

// QueryDocuments queries documents in a collection
func (a *App) QueryDocuments(sessionID string, req service.QueryRequest) ([]service.DocumentResponse, error) {
	dbService, err := a.getDBService(sessionID)
	if err != nil {
		return nil, err
	}
	return dbService.QueryDocuments(req)
}

// CreateIndex creates an index on a field in a collection
func (a *App) CreateIndex(sessionID, dbName, collName, field string) error {
	dbService, err := a.getDBService(sessionID)
	if err != nil {
		return err
	}
	return dbService.CreateIndex(dbName, collName, field)
}

// GetDatabaseStats returns statistics about a database
func (a *App) GetDatabaseStats(sessionID, dbName string) (map[string]interface{}, error) {
	dbService, err := a.getDBService(sessionID)
	if err != nil {
		return nil, err
	}
	return dbService.GetDatabaseStats(dbName)
}

// Advanced Query operations

// AdvancedQuery executes an advanced query with filters, sorting, and pagination
func (a *App) AdvancedQuery(sessionID string, req service.AdvancedQueryRequest) ([]service.DocumentResponse, error) {
	dbService, err := a.getDBService(sessionID)
	if err != nil {
		return nil, err
	}
	return dbService.AdvancedQuery(req)
}

// CountDocuments counts documents matching the query
func (a *App) CountDocuments(sessionID string, req service.AdvancedQueryRequest) (int, error) {
	dbService, err := a.getDBService(sessionID)
	if err != nil {
		return 0, err
	}
	return dbService.CountDocuments(req)
}

// Import/Export operations

// ExportData exports data from a collection
func (a *App) ExportData(sessionID string, req service.ExportRequest) error {
	dbService, err := a.getDBService(sessionID)
	if err != nil {
		return err
	}
	return dbService.ExportData(req)
}

// ImportData imports data into a collection
func (a *App) ImportData(sessionID string, req service.ImportRequest) (*engine.ImportResult, error) {
	dbService, err := a.getDBService(sessionID)
	if err != nil {
		return nil, err
	}
	return dbService.ImportData(req)
}

// ImportDataFromContent imports data from file content
func (a *App) ImportDataFromContent(sessionID, database, collection, content, format string, createCollection bool) (*engine.ImportResult, error) {
	dbService, err := a.getDBService(sessionID)
	if err != nil {
		return nil, err
	}
	return dbService.ImportDataFromContent(database, collection, content, format, createCollection)
}

// GetSupportedFormats returns supported import/export formats
func (a *App) GetSupportedFormats(sessionID string) ([]string, error) {
	dbService, err := a.getDBService(sessionID)
	if err != nil {
		return nil, err
	}
	return dbService.GetSupportedFormats(), nil
}

// Backup operations

// CreateBackup creates a backup of a database
func (a *App) CreateBackup(sessionID string, req service.BackupRequest) (*engine.BackupInfo, error) {
	dbService, err := a.getDBService(sessionID)
	if err != nil {
		return nil, err
	}
	return dbService.CreateBackup(req)
}

// RestoreBackup restores a database from backup
func (a *App) RestoreBackup(sessionID string, req service.RestoreRequest) error {
	dbService, err := a.getDBService(sessionID)
	if err != nil {
		return err
	}
	return dbService.RestoreBackup(req)
}

// ListBackups lists all available backups
func (a *App) ListBackups(sessionID string) ([]engine.BackupInfo, error) {
	dbService, err := a.getDBService(sessionID)
	if err != nil {
		return nil, err
	}
	return dbService.ListBackups()
}

// Database maintenance operations

// CompactDatabase compacts a database to optimize storage
func (a *App) CompactDatabase(sessionID, dbName string) error {
	dbService, err := a.getDBService(sessionID)
	if err != nil {
		return err
	}
	return dbService.CompactDatabase(dbName)
}

// GetDetailedDatabaseStats returns detailed statistics about a database
func (a *App) GetDetailedDatabaseStats(sessionID, dbName string) (*engine.DatabaseStats, error) {
	dbService, err := a.getDBService(sessionID)
	if err != nil {
		return nil, err
	}
	return dbService.GetDetailedDatabaseStats(dbName)
}
