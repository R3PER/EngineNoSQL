package engine

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ImportExportManager handles data import and export operations
type ImportExportManager struct {
	engine *Engine
}

// ImportFormat represents supported import/export formats
type ImportFormat string

const (
	FormatJSON = "json"
	FormatCSV  = "csv"
	FormatSQL  = "sql"
)

// ExportOptions contains options for export operations
type ExportOptions struct {
	Format     ImportFormat  `json:"format"`
	Collection string        `json:"collection"`
	Query      *QueryBuilder `json:"-"`
	FilePath   string        `json:"file_path"`
}

// ImportOptions contains options for import operations
type ImportOptions struct {
	Format           ImportFormat `json:"format"`
	Collection       string       `json:"collection"`
	FilePath         string       `json:"file_path"`
	CreateCollection bool         `json:"create_collection"`
	OverwriteData    bool         `json:"overwrite_data"`
	IDField          string       `json:"id_field"` // Field to use as document ID
}

// ImportResult contains results of import operation
type ImportResult struct {
	Imported int      `json:"imported"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors"`
}

// NewImportExportManager creates a new import/export manager
func NewImportExportManager(engine *Engine) *ImportExportManager {
	return &ImportExportManager{
		engine: engine,
	}
}

// ExportData exports data from a collection to a file
func (iem *ImportExportManager) ExportData(dbName string, options ExportOptions) error {
	db, err := iem.engine.GetDatabase(dbName)
	if err != nil {
		return fmt.Errorf("failed to get database: %v", err)
	}

	collection, err := db.GetCollection(options.Collection)
	if err != nil {
		return fmt.Errorf("failed to get collection: %v", err)
	}

	// Get documents to export
	var documents []*Document
	if options.Query != nil {
		documents, err = options.Query.Execute()
		if err != nil {
			return fmt.Errorf("failed to execute query: %v", err)
		}
	} else {
		documents = collection.GetAll()
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(options.FilePath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	switch options.Format {
	case FormatJSON:
		return iem.exportJSON(documents, options.FilePath)
	case FormatCSV:
		return iem.exportCSV(documents, options.FilePath)
	case FormatSQL:
		return iem.exportSQL(documents, options.Collection, options.FilePath)
	default:
		return fmt.Errorf("unsupported export format: %s", options.Format)
	}
}

// ImportData imports data from a file into a collection
func (iem *ImportExportManager) ImportData(dbName string, options ImportOptions) (*ImportResult, error) {
	db, err := iem.engine.GetDatabase(dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %v", err)
	}

	// Create collection if it doesn't exist and requested
	collection, err := db.GetCollection(options.Collection)
	if err != nil && options.CreateCollection {
		if err := db.CreateCollection(options.Collection); err != nil {
			return nil, fmt.Errorf("failed to create collection: %v", err)
		}
		collection, err = db.GetCollection(options.Collection)
		if err != nil {
			return nil, fmt.Errorf("failed to get created collection: %v", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("collection not found: %v", err)
	}

	// Clear existing data if requested
	if options.OverwriteData {
		collection.mutex.Lock()
		collection.Documents = make(map[string]*Document)
		collection.mutex.Unlock()
	}

	switch options.Format {
	case FormatJSON:
		return iem.importJSON(collection, options)
	case FormatCSV:
		return iem.importCSV(collection, options)
	case FormatSQL:
		return iem.importSQL(collection, options)
	default:
		return nil, fmt.Errorf("unsupported import format: %s", options.Format)
	}
}

// exportJSON exports documents to JSON format
func (iem *ImportExportManager) exportJSON(documents []*Document, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	// Create export structure
	exportData := struct {
		ExportedAt time.Time   `json:"exported_at"`
		Count      int         `json:"count"`
		Documents  []*Document `json:"documents"`
	}{
		ExportedAt: time.Now(),
		Count:      len(documents),
		Documents:  documents,
	}

	return encoder.Encode(exportData)
}

// exportCSV exports documents to CSV format
func (iem *ImportExportManager) exportCSV(documents []*Document, filePath string) error {
	if len(documents) == 0 {
		return fmt.Errorf("no documents to export")
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Collect all unique field names
	fieldSet := make(map[string]bool)
	fieldSet["_id"] = true
	fieldSet["created_at"] = true
	fieldSet["updated_at"] = true

	for _, doc := range documents {
		for field := range doc.Data {
			fieldSet[field] = true
		}
	}

	// Create header row
	var headers []string
	headers = append(headers, "_id", "created_at", "updated_at")
	for field := range fieldSet {
		if field != "_id" && field != "created_at" && field != "updated_at" {
			headers = append(headers, field)
		}
	}

	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %v", err)
	}

	// Write data rows
	for _, doc := range documents {
		var row []string
		row = append(row, doc.ID)
		row = append(row, doc.CreatedAt.Format(time.RFC3339))
		row = append(row, doc.UpdatedAt.Format(time.RFC3339))

		for _, field := range headers[3:] { // Skip _id, created_at, updated_at
			if value, exists := doc.Data[field]; exists {
				row = append(row, fmt.Sprintf("%v", value))
			} else {
				row = append(row, "")
			}
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %v", err)
		}
	}

	return nil
}

// exportSQL exports documents to SQL INSERT statements
func (iem *ImportExportManager) exportSQL(documents []*Document, tableName, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Write header
	fmt.Fprintf(file, "-- SQL Export for table: %s\n", tableName)
	fmt.Fprintf(file, "-- Generated at: %s\n\n", time.Now().Format(time.RFC3339))

	// Write CREATE TABLE statement (basic structure)
	fmt.Fprintf(file, "CREATE TABLE IF NOT EXISTS %s (\n", tableName)
	fmt.Fprintf(file, "    id VARCHAR(255) PRIMARY KEY,\n")
	fmt.Fprintf(file, "    data JSON,\n")
	fmt.Fprintf(file, "    created_at TIMESTAMP,\n")
	fmt.Fprintf(file, "    updated_at TIMESTAMP\n")
	fmt.Fprintf(file, ");\n\n")

	// Write INSERT statements
	for _, doc := range documents {
		dataJSON, _ := json.Marshal(doc.Data)
		fmt.Fprintf(file, "INSERT INTO %s (id, data, created_at, updated_at) VALUES ('%s', '%s', '%s', '%s');\n",
			tableName,
			strings.ReplaceAll(doc.ID, "'", "''"),
			strings.ReplaceAll(string(dataJSON), "'", "''"),
			doc.CreatedAt.Format("2006-01-02 15:04:05"),
			doc.UpdatedAt.Format("2006-01-02 15:04:05"))
	}

	return nil
}

// importJSON imports documents from JSON format
func (iem *ImportExportManager) importJSON(collection *Collection, options ImportOptions) (*ImportResult, error) {
	file, err := os.Open(options.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	result := &ImportResult{
		Errors: make([]string, 0),
	}

	// Parse JSON content
	var jsonData interface{}
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %v", err)
	}

	// Try different import strategies
	documents, err := iem.parseJsonData(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON data: %v", err)
	}

	// Import documents
	for i, docData := range documents {
		var docID string
		if options.IDField != "" && docData[options.IDField] != nil {
			docID = fmt.Sprintf("%v", docData[options.IDField])
			delete(docData, options.IDField) // Remove ID from data
		} else {
			docID = fmt.Sprintf("imported_%d_%d", time.Now().UnixNano(), i)
		}

		if err := collection.Insert(docID, docData); err != nil {
			result.Skipped++
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to insert %s: %v", docID, err))
		} else {
			result.Imported++
		}
	}

	return result, nil
}

// parseJsonData parses various JSON formats and extracts documents
func (iem *ImportExportManager) parseJsonData(jsonData interface{}) ([]map[string]interface{}, error) {
	switch data := jsonData.(type) {
	case map[string]interface{}:
		// Check if it's export format with documents array
		if documents, ok := data["documents"]; ok {
			return iem.parseDocumentsArray(documents)
		}
		// Check if it's a single document
		return []map[string]interface{}{data}, nil
	case []interface{}:
		// Array of documents
		var documents []map[string]interface{}
		for _, item := range data {
			if doc, ok := item.(map[string]interface{}); ok {
				documents = append(documents, doc)
			}
		}
		return documents, nil
	default:
		return nil, fmt.Errorf("unsupported JSON format")
	}
}

// parseDocumentsArray parses the documents array from export format
func (iem *ImportExportManager) parseDocumentsArray(documents interface{}) ([]map[string]interface{}, error) {
	switch docs := documents.(type) {
	case []interface{}:
		var result []map[string]interface{}
		for _, doc := range docs {
			if docMap, ok := doc.(map[string]interface{}); ok {
				// Extract actual data from document structure
				if data, exists := docMap["data"]; exists {
					if dataMap, ok := data.(map[string]interface{}); ok {
						result = append(result, dataMap)
					}
				} else {
					// If no nested data field, use the document as is
					result = append(result, docMap)
				}
			}
		}
		return result, nil
	default:
		return nil, fmt.Errorf("documents field is not an array")
	}
}

// importCSV imports documents from CSV format
func (iem *ImportExportManager) importCSV(collection *Collection, options ImportOptions) (*ImportResult, error) {
	file, err := os.Open(options.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %v", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must have at least header and one data row")
	}

	headers := records[0]
	result := &ImportResult{
		Errors: make([]string, 0),
	}

	// Find ID column
	idColumn := -1
	if options.IDField != "" {
		for i, header := range headers {
			if header == options.IDField {
				idColumn = i
				break
			}
		}
	}

	// Process data rows
	for rowIndex, record := range records[1:] {
		if len(record) != len(headers) {
			result.Skipped++
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: column count mismatch", rowIndex+2))
			continue
		}

		// Create document data
		docData := make(map[string]interface{})
		var docID string

		for i, value := range record {
			header := headers[i]

			if i == idColumn {
				docID = value
				continue
			}

			// Skip system fields in data
			if header == "_id" || header == "created_at" || header == "updated_at" {
				if header == "_id" && docID == "" {
					docID = value
				}
				continue
			}

			// Try to parse as number first, then as string
			if numValue, err := strconv.ParseFloat(value, 64); err == nil {
				if numValue == float64(int64(numValue)) {
					docData[header] = int64(numValue)
				} else {
					docData[header] = numValue
				}
			} else if boolValue, err := strconv.ParseBool(value); err == nil {
				docData[header] = boolValue
			} else {
				docData[header] = value
			}
		}

		// Generate ID if not provided
		if docID == "" {
			docID = fmt.Sprintf("csv_import_%d_%d", time.Now().Unix(), rowIndex)
		}

		// Insert document
		if err := collection.Insert(docID, docData); err != nil {
			result.Skipped++
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to insert %s: %v", docID, err))
		} else {
			result.Imported++
		}
	}

	return result, nil
}

// importSQL imports documents from SQL file (basic implementation)
func (iem *ImportExportManager) importSQL(collection *Collection, options ImportOptions) (*ImportResult, error) {
	file, err := os.Open(options.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	result := &ImportResult{
		Errors: make([]string, 0),
	}

	// Parse INSERT statements (basic regex-based parsing)
	lines := strings.Split(string(content), "\n")
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(strings.ToUpper(line), "INSERT") {
			continue
		}

		// Try to extract values (very basic parsing)
		if err := iem.parseInsertStatement(collection, line, result); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Line %d: %v", lineNum+1, err))
		}
	}

	return result, nil
}

// parseInsertStatement parses a basic INSERT statement
func (iem *ImportExportManager) parseInsertStatement(collection *Collection, statement string, result *ImportResult) error {
	// This is a very basic parser - in production, you'd want a proper SQL parser
	// For now, we'll handle our own export format

	valuesIndex := strings.Index(strings.ToUpper(statement), "VALUES")
	if valuesIndex == -1 {
		result.Skipped++
		return fmt.Errorf("invalid INSERT statement")
	}

	valuesStr := statement[valuesIndex+6:]
	valuesStr = strings.TrimSpace(valuesStr)

	// Remove parentheses and semicolon
	valuesStr = strings.Trim(valuesStr, "()")
	valuesStr = strings.TrimSuffix(valuesStr, ";")

	// Split by comma (basic - doesn't handle quoted commas properly)
	parts := strings.Split(valuesStr, "', '")
	if len(parts) < 4 {
		result.Skipped++
		return fmt.Errorf("insufficient values in INSERT statement")
	}

	// Clean up quotes
	for i := range parts {
		parts[i] = strings.Trim(parts[i], "'")
	}

	docID := parts[0]
	dataJSON := parts[1]

	// Parse JSON data
	var docData map[string]interface{}
	if err := json.Unmarshal([]byte(dataJSON), &docData); err != nil {
		result.Skipped++
		return fmt.Errorf("failed to parse JSON data: %v", err)
	}

	// Insert document
	if err := collection.Insert(docID, docData); err != nil {
		result.Skipped++
		return fmt.Errorf("failed to insert document: %v", err)
	}

	result.Imported++
	return nil
}

// ImportDataFromContent imports data from file content string
func (iem *ImportExportManager) ImportDataFromContent(dbName, collectionName, content, format string, createCollection bool) (*ImportResult, error) {
	db, err := iem.engine.GetDatabase(dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %v", err)
	}

	// Create collection if it doesn't exist and requested
	collection, err := db.GetCollection(collectionName)
	if err != nil && createCollection {
		if err := db.CreateCollection(collectionName); err != nil {
			return nil, fmt.Errorf("failed to create collection: %v", err)
		}
		collection, err = db.GetCollection(collectionName)
		if err != nil {
			return nil, fmt.Errorf("failed to get created collection: %v", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("collection not found: %v", err)
	}

	switch ImportFormat(format) {
	case FormatJSON:
		return iem.importJSONFromContent(collection, content)
	case FormatCSV:
		return iem.importCSVFromContent(collection, content)
	default:
		return nil, fmt.Errorf("unsupported import format: %s", format)
	}
}

// importJSONFromContent imports JSON data from content string
func (iem *ImportExportManager) importJSONFromContent(collection *Collection, content string) (*ImportResult, error) {
	result := &ImportResult{
		Errors: make([]string, 0),
	}

	// Parse JSON content
	var jsonData interface{}
	if err := json.Unmarshal([]byte(content), &jsonData); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %v", err)
	}

	// Try different import strategies
	documents, err := iem.parseJsonData(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON data: %v", err)
	}

	// Import documents
	for i, docData := range documents {
		docID := fmt.Sprintf("imported_%d_%d", time.Now().UnixNano(), i)

		if err := collection.Insert(docID, docData); err != nil {
			result.Skipped++
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to insert %s: %v", docID, err))
		} else {
			result.Imported++
		}
	}

	return result, nil
}

// importCSVFromContent imports CSV data from content string
func (iem *ImportExportManager) importCSVFromContent(collection *Collection, content string) (*ImportResult, error) {
	reader := csv.NewReader(strings.NewReader(content))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %v", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV content must have at least header and one data row")
	}

	headers := records[0]
	result := &ImportResult{
		Errors: make([]string, 0),
	}

	// Process data rows
	for rowIndex, record := range records[1:] {
		if len(record) != len(headers) {
			result.Skipped++
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: column count mismatch", rowIndex+2))
			continue
		}

		// Create document data
		docData := make(map[string]interface{})
		docID := fmt.Sprintf("csv_imported_%d_%d", time.Now().UnixNano(), rowIndex)

		for i, value := range record {
			header := headers[i]

			// Skip system fields in data
			if header == "_id" || header == "created_at" || header == "updated_at" {
				if header == "_id" {
					docID = value
				}
				continue
			}

			// Try to parse as number first, then as string
			if numValue, err := strconv.ParseFloat(value, 64); err == nil {
				if numValue == float64(int64(numValue)) {
					docData[header] = int64(numValue)
				} else {
					docData[header] = numValue
				}
			} else if boolValue, err := strconv.ParseBool(value); err == nil {
				docData[header] = boolValue
			} else {
				docData[header] = value
			}
		}

		// Insert document
		if err := collection.Insert(docID, docData); err != nil {
			result.Skipped++
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to insert %s: %v", docID, err))
		} else {
			result.Imported++
		}
	}

	return result, nil
}

// GetSupportedFormats returns list of supported import/export formats
func (iem *ImportExportManager) GetSupportedFormats() []string {
	return []string{string(FormatJSON), string(FormatCSV), string(FormatSQL)}
}
