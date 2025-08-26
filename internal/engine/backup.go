package engine

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// BackupManager handles database backups and restoration
type BackupManager struct {
	engine    *Engine
	backupDir string
}

// BackupInfo contains information about a backup
type BackupInfo struct {
	Name      string    `json:"name"`
	Database  string    `json:"database"`
	Timestamp time.Time `json:"timestamp"`
	Size      int64     `json:"size"`
	Path      string    `json:"path"`
}

// NewBackupManager creates a new backup manager
func NewBackupManager(engine *Engine, backupDir string) *BackupManager {
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create backup directory: %v", err))
	}

	return &BackupManager{
		engine:    engine,
		backupDir: backupDir,
	}
}

// CreateBackup creates a compressed backup of a database
func (bm *BackupManager) CreateBackup(dbName, backupName string) (*BackupInfo, error) {
	db, err := bm.engine.GetDatabase(dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %v", err)
	}

	// Create backup filename with timestamp
	timestamp := time.Now()
	filename := fmt.Sprintf("%s_%s_%s.tar.gz", dbName, backupName, timestamp.Format("20060102_150405"))
	backupPath := filepath.Join(bm.backupDir, filename)

	// Create compressed tar file
	file, err := os.Create(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup file: %v", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add database file to tar
	dbData, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal database: %v", err)
	}

	header := &tar.Header{
		Name: dbName + ".enosql",
		Mode: 0644,
		Size: int64(len(dbData)),
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		return nil, fmt.Errorf("failed to write tar header: %v", err)
	}

	if _, err := tarWriter.Write(dbData); err != nil {
		return nil, fmt.Errorf("failed to write database data: %v", err)
	}

	// Add metadata file
	metadata := map[string]interface{}{
		"database":  dbName,
		"backup":    backupName,
		"timestamp": timestamp,
		"version":   "1.0",
	}

	metaData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %v", err)
	}

	metaHeader := &tar.Header{
		Name: "metadata.json",
		Mode: 0644,
		Size: int64(len(metaData)),
	}

	if err := tarWriter.WriteHeader(metaHeader); err != nil {
		return nil, fmt.Errorf("failed to write metadata header: %v", err)
	}

	if _, err := tarWriter.Write(metaData); err != nil {
		return nil, fmt.Errorf("failed to write metadata: %v", err)
	}

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}

	backupInfo := &BackupInfo{
		Name:      backupName,
		Database:  dbName,
		Timestamp: timestamp,
		Size:      fileInfo.Size(),
		Path:      backupPath,
	}

	return backupInfo, nil
}

// RestoreBackup restores a database from a backup
func (bm *BackupManager) RestoreBackup(backupPath, newDbName string) error {
	// Open backup file
	file, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %v", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	var dbData []byte
	var metadata map[string]interface{}

	// Read tar contents
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %v", err)
		}

		switch header.Name {
		case "metadata.json":
			data, err := io.ReadAll(tarReader)
			if err != nil {
				return fmt.Errorf("failed to read metadata: %v", err)
			}
			if err := json.Unmarshal(data, &metadata); err != nil {
				return fmt.Errorf("failed to unmarshal metadata: %v", err)
			}

		default:
			if filepath.Ext(header.Name) == ".enosql" {
				data, err := io.ReadAll(tarReader)
				if err != nil {
					return fmt.Errorf("failed to read database data: %v", err)
				}
				dbData = data
			}
		}
	}

	if len(dbData) == 0 {
		return fmt.Errorf("no database data found in backup")
	}

	// Unmarshal database
	var db Database
	if err := json.Unmarshal(dbData, &db); err != nil {
		return fmt.Errorf("failed to unmarshal database: %v", err)
	}

	// Update database name and path
	db.Name = newDbName
	db.Path = filepath.Join(bm.engine.dataDir, newDbName+".enosql")

	// Initialize mutexes
	for _, collection := range db.Collections {
		collection.mutex = sync.RWMutex{}
	}

	// Save to engine
	bm.engine.mutex.Lock()
	bm.engine.databases[newDbName] = &db
	bm.engine.mutex.Unlock()

	// Save to disk
	return bm.engine.saveDatabase(&db)
}

// ListBackups lists all available backups
func (bm *BackupManager) ListBackups() ([]BackupInfo, error) {
	var backups []BackupInfo

	files, err := filepath.Glob(filepath.Join(bm.backupDir, "*.tar.gz"))
	if err != nil {
		return nil, fmt.Errorf("failed to list backup files: %v", err)
	}

	for _, filePath := range files {
		info, err := bm.getBackupInfo(filePath)
		if err != nil {
			continue // Skip invalid backups
		}
		backups = append(backups, *info)
	}

	return backups, nil
}

// getBackupInfo extracts backup information from a backup file
func (bm *BackupManager) getBackupInfo(filePath string) (*BackupInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	// Look for metadata file
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if header.Name == "metadata.json" {
			data, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, err
			}

			var metadata map[string]interface{}
			if err := json.Unmarshal(data, &metadata); err != nil {
				return nil, err
			}

			timestampStr, ok := metadata["timestamp"].(string)
			if !ok {
				return nil, fmt.Errorf("invalid timestamp in metadata")
			}

			timestamp, err := time.Parse(time.RFC3339Nano, timestampStr)
			if err != nil {
				return nil, err
			}

			return &BackupInfo{
				Name:      metadata["backup"].(string),
				Database:  metadata["database"].(string),
				Timestamp: timestamp,
				Size:      fileInfo.Size(),
				Path:      filePath,
			}, nil
		}
	}

	return nil, fmt.Errorf("no metadata found in backup")
}

// DeleteBackup deletes a backup file
func (bm *BackupManager) DeleteBackup(backupPath string) error {
	return os.Remove(backupPath)
}

// CompactDatabase optimizes database storage by removing gaps and reorganizing data
func (e *Engine) CompactDatabase(dbName string) error {
	db, err := e.GetDatabase(dbName)
	if err != nil {
		return err
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	// Rebuild indexes for all collections
	for _, collection := range db.Collections {
		collection.mutex.Lock()

		// Clear and rebuild indexes
		for field := range collection.Indexes {
			index := Index{
				Field:  field,
				Values: make(map[string]string),
			}

			// Rebuild index from existing documents
			for _, doc := range collection.Documents {
				if value, exists := doc.Data[field]; exists {
					valueStr := fmt.Sprintf("%v", value)
					index.Values[valueStr] = doc.ID
				}
			}

			collection.Indexes[field] = index
		}

		collection.mutex.Unlock()
	}

	// Save the compacted database
	return e.saveDatabase(db)
}

// DatabaseStats provides detailed statistics about a database
type DatabaseStats struct {
	Name            string                     `json:"name"`
	Collections     int                        `json:"collections"`
	TotalDocuments  int                        `json:"total_documents"`
	TotalIndexes    int                        `json:"total_indexes"`
	SizeOnDisk      int64                      `json:"size_on_disk"`
	CollectionStats map[string]CollectionStats `json:"collection_stats"`
}

// CollectionStats provides statistics about a collection
type CollectionStats struct {
	Name            string             `json:"name"`
	DocumentCount   int                `json:"document_count"`
	IndexCount      int                `json:"index_count"`
	AvgDocSize      float64            `json:"avg_doc_size"`
	FieldTypes      map[string]string  `json:"field_types"`
	IndexEfficiency map[string]float64 `json:"index_efficiency"`
}

// GetDatabaseStats returns detailed statistics about a database
func (e *Engine) GetDatabaseStats(dbName string) (*DatabaseStats, error) {
	db, err := e.GetDatabase(dbName)
	if err != nil {
		return nil, err
	}

	db.mutex.RLock()
	defer db.mutex.RUnlock()

	stats := &DatabaseStats{
		Name:            dbName,
		Collections:     len(db.Collections),
		CollectionStats: make(map[string]CollectionStats),
	}

	// Get file size
	if fileInfo, err := os.Stat(db.Path); err == nil {
		stats.SizeOnDisk = fileInfo.Size()
	}

	totalDocs := 0
	totalIndexes := 0

	for _, collection := range db.Collections {
		collection.mutex.RLock()

		collStats := CollectionStats{
			Name:            collection.Name,
			DocumentCount:   len(collection.Documents),
			IndexCount:      len(collection.Indexes),
			FieldTypes:      make(map[string]string),
			IndexEfficiency: make(map[string]float64),
		}

		// Calculate average document size
		totalSize := 0
		fieldTypeCounts := make(map[string]map[string]int)

		for _, doc := range collection.Documents {
			docData, _ := json.Marshal(doc)
			totalSize += len(docData)

			// Analyze field types
			for field, value := range doc.Data {
				if fieldTypeCounts[field] == nil {
					fieldTypeCounts[field] = make(map[string]int)
				}
				fieldType := getValueType(value)
				fieldTypeCounts[field][fieldType]++
			}
		}

		if len(collection.Documents) > 0 {
			collStats.AvgDocSize = float64(totalSize) / float64(len(collection.Documents))
		}

		// Determine dominant field types
		for field, typeCounts := range fieldTypeCounts {
			maxCount := 0
			dominantType := "unknown"
			for fieldType, count := range typeCounts {
				if count > maxCount {
					maxCount = count
					dominantType = fieldType
				}
			}
			collStats.FieldTypes[field] = dominantType
		}

		// Calculate index efficiency
		for field, index := range collection.Indexes {
			if len(collection.Documents) > 0 {
				efficiency := float64(len(index.Values)) / float64(len(collection.Documents))
				collStats.IndexEfficiency[field] = efficiency
			}
		}

		stats.CollectionStats[collection.Name] = collStats
		totalDocs += len(collection.Documents)
		totalIndexes += len(collection.Indexes)

		collection.mutex.RUnlock()
	}

	stats.TotalDocuments = totalDocs
	stats.TotalIndexes = totalIndexes

	return stats, nil
}
