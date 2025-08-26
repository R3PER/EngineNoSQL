package engine

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

// QueryBuilder provides advanced query capabilities
type QueryBuilder struct {
	collection *Collection
	filters    []Filter
	sortBy     string
	sortOrder  int // 1 for ascending, -1 for descending
	limit      int
	skip       int
}

// Filter represents a query filter
type Filter struct {
	Field    string
	Operator string
	Value    interface{}
}

// Operators
const (
	OpEqual              = "$eq"
	OpNotEqual           = "$ne"
	OpGreaterThan        = "$gt"
	OpGreaterThanOrEqual = "$gte"
	OpLessThan           = "$lt"
	OpLessThanOrEqual    = "$lte"
	OpIn                 = "$in"
	OpNotIn              = "$nin"
	OpRegex              = "$regex"
	OpExists             = "$exists"
	OpType               = "$type"
	OpSize               = "$size"
)

// NewQueryBuilder creates a new query builder for a collection
func (c *Collection) NewQuery() *QueryBuilder {
	return &QueryBuilder{
		collection: c,
		filters:    make([]Filter, 0),
		sortOrder:  1,
		limit:      0,
		skip:       0,
	}
}

// Where adds a filter to the query
func (qb *QueryBuilder) Where(field, operator string, value interface{}) *QueryBuilder {
	qb.filters = append(qb.filters, Filter{
		Field:    field,
		Operator: operator,
		Value:    value,
	})
	return qb
}

// Equals adds an equality filter
func (qb *QueryBuilder) Equals(field string, value interface{}) *QueryBuilder {
	return qb.Where(field, OpEqual, value)
}

// GreaterThan adds a greater than filter
func (qb *QueryBuilder) GreaterThan(field string, value interface{}) *QueryBuilder {
	return qb.Where(field, OpGreaterThan, value)
}

// LessThan adds a less than filter
func (qb *QueryBuilder) LessThan(field string, value interface{}) *QueryBuilder {
	return qb.Where(field, OpLessThan, value)
}

// In adds an "in" filter for arrays
func (qb *QueryBuilder) In(field string, values []interface{}) *QueryBuilder {
	return qb.Where(field, OpIn, values)
}

// Regex adds a regex filter
func (qb *QueryBuilder) Regex(field string, pattern string) *QueryBuilder {
	return qb.Where(field, OpRegex, pattern)
}

// Exists checks if field exists
func (qb *QueryBuilder) Exists(field string, exists bool) *QueryBuilder {
	return qb.Where(field, OpExists, exists)
}

// Sort sets the sort field and order
func (qb *QueryBuilder) Sort(field string, ascending bool) *QueryBuilder {
	qb.sortBy = field
	if ascending {
		qb.sortOrder = 1
	} else {
		qb.sortOrder = -1
	}
	return qb
}

// Limit sets the maximum number of results
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Skip sets the number of results to skip
func (qb *QueryBuilder) Skip(skip int) *QueryBuilder {
	qb.skip = skip
	return qb
}

// Execute runs the query and returns matching documents
func (qb *QueryBuilder) Execute() ([]*Document, error) {
	qb.collection.mutex.RLock()
	defer qb.collection.mutex.RUnlock()

	var results []*Document

	// Filter documents
	for _, doc := range qb.collection.Documents {
		if qb.matchesFilters(doc) {
			results = append(results, doc)
		}
	}

	// Sort results
	if qb.sortBy != "" {
		qb.sortDocuments(results)
	}

	// Apply skip
	if qb.skip > 0 && qb.skip < len(results) {
		results = results[qb.skip:]
	} else if qb.skip >= len(results) {
		results = []*Document{}
	}

	// Apply limit
	if qb.limit > 0 && qb.limit < len(results) {
		results = results[:qb.limit]
	}

	return results, nil
}

// Count returns the number of documents matching the query
func (qb *QueryBuilder) Count() (int, error) {
	qb.collection.mutex.RLock()
	defer qb.collection.mutex.RUnlock()

	count := 0
	for _, doc := range qb.collection.Documents {
		if qb.matchesFilters(doc) {
			count++
		}
	}

	return count, nil
}

// matchesFilters checks if a document matches all filters
func (qb *QueryBuilder) matchesFilters(doc *Document) bool {
	for _, filter := range qb.filters {
		if !qb.matchesFilter(doc, filter) {
			return false
		}
	}
	return true
}

// matchesFilter checks if a document matches a single filter
func (qb *QueryBuilder) matchesFilter(doc *Document, filter Filter) bool {
	fieldValue, exists := doc.Data[filter.Field]

	switch filter.Operator {
	case OpEqual:
		return exists && compareValues(fieldValue, filter.Value) == 0

	case OpNotEqual:
		return !exists || compareValues(fieldValue, filter.Value) != 0

	case OpGreaterThan:
		return exists && compareValues(fieldValue, filter.Value) > 0

	case OpGreaterThanOrEqual:
		return exists && compareValues(fieldValue, filter.Value) >= 0

	case OpLessThan:
		return exists && compareValues(fieldValue, filter.Value) < 0

	case OpLessThanOrEqual:
		return exists && compareValues(fieldValue, filter.Value) <= 0

	case OpIn:
		if !exists {
			return false
		}
		values, ok := filter.Value.([]interface{})
		if !ok {
			return false
		}
		for _, value := range values {
			if compareValues(fieldValue, value) == 0 {
				return true
			}
		}
		return false

	case OpNotIn:
		if !exists {
			return true
		}
		values, ok := filter.Value.([]interface{})
		if !ok {
			return true
		}
		for _, value := range values {
			if compareValues(fieldValue, value) == 0 {
				return false
			}
		}
		return true

	case OpRegex:
		if !exists {
			return false
		}
		pattern, ok := filter.Value.(string)
		if !ok {
			return false
		}
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return false
		}
		fieldStr := fmt.Sprintf("%v", fieldValue)
		return regex.MatchString(fieldStr)

	case OpExists:
		expected, ok := filter.Value.(bool)
		if !ok {
			return false
		}
		return exists == expected

	case OpType:
		if !exists {
			return false
		}
		expectedType, ok := filter.Value.(string)
		if !ok {
			return false
		}
		return getValueType(fieldValue) == expectedType

	case OpSize:
		if !exists {
			return false
		}
		expectedSize, ok := filter.Value.(int)
		if !ok {
			return false
		}
		return getValueSize(fieldValue) == expectedSize

	default:
		return false
	}
}

// compareValues compares two values and returns -1, 0, or 1
func compareValues(a, b interface{}) int {
	// Convert to strings for comparison
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)

	// Try numeric comparison first
	if aNum, aErr := strconv.ParseFloat(aStr, 64); aErr == nil {
		if bNum, bErr := strconv.ParseFloat(bStr, 64); bErr == nil {
			if aNum < bNum {
				return -1
			} else if aNum > bNum {
				return 1
			} else {
				return 0
			}
		}
	}

	// String comparison
	if aStr < bStr {
		return -1
	} else if aStr > bStr {
		return 1
	} else {
		return 0
	}
}

// getValueType returns the type of a value as a string
func getValueType(value interface{}) string {
	if value == nil {
		return "null"
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "int"
	case reflect.Float32, reflect.Float64:
		return "double"
	case reflect.Bool:
		return "bool"
	case reflect.Array, reflect.Slice:
		return "array"
	case reflect.Map:
		return "object"
	default:
		return "unknown"
	}
}

// getValueSize returns the size of a value (for arrays, strings, etc.)
func getValueSize(value interface{}) int {
	if value == nil {
		return 0
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return len(v.String())
	case reflect.Array, reflect.Slice:
		return v.Len()
	case reflect.Map:
		return v.Len()
	default:
		return 0
	}
}

// sortDocuments sorts documents by the specified field
func (qb *QueryBuilder) sortDocuments(docs []*Document) {
	if qb.sortBy == "" {
		return
	}

	// Simple bubble sort implementation
	n := len(docs)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			shouldSwap := false

			val1, exists1 := docs[j].Data[qb.sortBy]
			val2, exists2 := docs[j+1].Data[qb.sortBy]

			if !exists1 && exists2 {
				shouldSwap = qb.sortOrder == 1
			} else if exists1 && !exists2 {
				shouldSwap = qb.sortOrder == -1
			} else if exists1 && exists2 {
				cmp := compareValues(val1, val2)
				shouldSwap = (qb.sortOrder == 1 && cmp > 0) || (qb.sortOrder == -1 && cmp < 0)
			}

			if shouldSwap {
				docs[j], docs[j+1] = docs[j+1], docs[j]
			}
		}
	}
}

// Aggregation functions

// Aggregate performs aggregation operations
func (c *Collection) Aggregate(pipeline []AggregationStage) ([]map[string]interface{}, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Convert documents to map format for aggregation
	var data []map[string]interface{}
	for _, doc := range c.Documents {
		item := make(map[string]interface{})
		item["_id"] = doc.ID
		item["created_at"] = doc.CreatedAt
		item["updated_at"] = doc.UpdatedAt
		for k, v := range doc.Data {
			item[k] = v
		}
		data = append(data, item)
	}

	// Process each stage in the pipeline
	for _, stage := range pipeline {
		var err error
		data, err = stage.Process(data)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

// AggregationStage represents a stage in an aggregation pipeline
type AggregationStage interface {
	Process(data []map[string]interface{}) ([]map[string]interface{}, error)
}

// MatchStage filters documents
type MatchStage struct {
	Filters []Filter
}

func (s *MatchStage) Process(data []map[string]interface{}) ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	for _, item := range data {
		matches := true
		for _, filter := range s.Filters {
			if !s.matchesFilter(item, filter) {
				matches = false
				break
			}
		}
		if matches {
			result = append(result, item)
		}
	}

	return result, nil
}

func (s *MatchStage) matchesFilter(item map[string]interface{}, filter Filter) bool {
	fieldValue, exists := item[filter.Field]

	switch filter.Operator {
	case OpEqual:
		return exists && compareValues(fieldValue, filter.Value) == 0
	case OpGreaterThan:
		return exists && compareValues(fieldValue, filter.Value) > 0
	case OpLessThan:
		return exists && compareValues(fieldValue, filter.Value) < 0
	// Add more operators as needed
	default:
		return false
	}
}

// GroupStage groups documents by specified fields
type GroupStage struct {
	ID     interface{}              // Grouping key
	Fields map[string]AggregateFunc // Fields to aggregate
}

type AggregateFunc struct {
	Operation string // sum, avg, count, max, min
	Field     string
}

func (s *GroupStage) Process(data []map[string]interface{}) ([]map[string]interface{}, error) {
	groups := make(map[string][]map[string]interface{})

	// Group documents
	for _, item := range data {
		groupKey := s.getGroupKey(item)
		groups[groupKey] = append(groups[groupKey], item)
	}

	// Calculate aggregations
	var result []map[string]interface{}
	for groupKey, groupData := range groups {
		groupResult := make(map[string]interface{})
		groupResult["_id"] = groupKey

		for fieldName, aggFunc := range s.Fields {
			value, err := s.calculateAggregation(groupData, aggFunc)
			if err != nil {
				return nil, err
			}
			groupResult[fieldName] = value
		}

		result = append(result, groupResult)
	}

	return result, nil
}

func (s *GroupStage) getGroupKey(item map[string]interface{}) string {
	if idStr, ok := s.ID.(string); ok {
		if value, exists := item[idStr]; exists {
			return fmt.Sprintf("%v", value)
		}
	}
	return "null"
}

func (s *GroupStage) calculateAggregation(data []map[string]interface{}, aggFunc AggregateFunc) (interface{}, error) {
	switch aggFunc.Operation {
	case "count":
		return len(data), nil

	case "sum":
		sum := 0.0
		for _, item := range data {
			if value, exists := item[aggFunc.Field]; exists {
				if num, err := strconv.ParseFloat(fmt.Sprintf("%v", value), 64); err == nil {
					sum += num
				}
			}
		}
		return sum, nil

	case "avg":
		sum := 0.0
		count := 0
		for _, item := range data {
			if value, exists := item[aggFunc.Field]; exists {
				if num, err := strconv.ParseFloat(fmt.Sprintf("%v", value), 64); err == nil {
					sum += num
					count++
				}
			}
		}
		if count > 0 {
			return sum / float64(count), nil
		}
		return 0, nil

	case "max":
		var max float64
		first := true
		for _, item := range data {
			if value, exists := item[aggFunc.Field]; exists {
				if num, err := strconv.ParseFloat(fmt.Sprintf("%v", value), 64); err == nil {
					if first || num > max {
						max = num
						first = false
					}
				}
			}
		}
		if first {
			return nil, nil
		}
		return max, nil

	case "min":
		var min float64
		first := true
		for _, item := range data {
			if value, exists := item[aggFunc.Field]; exists {
				if num, err := strconv.ParseFloat(fmt.Sprintf("%v", value), 64); err == nil {
					if first || num < min {
						min = num
						first = false
					}
				}
			}
		}
		if first {
			return nil, nil
		}
		return min, nil

	default:
		return nil, fmt.Errorf("unknown aggregation operation: %s", aggFunc.Operation)
	}
}
