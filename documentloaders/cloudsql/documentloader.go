package cloudsql

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/util/cloudsqlutil"
)

const (
	defaultMetadataJSONColumn = "langchain_metadata"
	defaultSchemaName         = "public"
)

// DocumentLoader is responsible for loading documents from a Postgres database.
type DocumentLoader struct {
	engine             cloudsqlutil.PostgresEngine
	query              string
	tableName          string
	schemaName         string
	contentColumns     []string
	metadataColumns    []string
	metadataJSONColumn string
	formatter          func(map[string]any, []string) string
}

// NewDocumentLoader creates a new DocumentLoader instance.
func NewDocumentLoader(ctx context.Context, engine cloudsqlutil.PostgresEngine, options ...DocumentLoaderOption) (*DocumentLoader, error) {
	documentLoader := &DocumentLoader{
		engine:     engine,
		schemaName: defaultSchemaName,
	}

	for _, opt := range options {
		opt(documentLoader)
	}

	if err := validateDocumentLoader(documentLoader); err != nil {
		return nil, err
	}

	if err := validateQuery(documentLoader.query); err != nil {
		return nil, err
	}

	columNames, err := documentLoader.getColumNames(ctx)
	if err != nil {
		return nil, err
	}

	if err = documentLoader.configureColumns(columNames); err != nil {
		return nil, err
	}

	if err = documentLoader.validateColumns(columNames); err != nil {
		return nil, err
	}

	return documentLoader, nil
}

// textFormatter formats row data into a text string.
func textFormatter(row map[string]any, contentColumns []string) string {
	var sb strings.Builder
	for _, column := range contentColumns {
		if val, ok := row[column]; ok {
			sb.WriteString(fmt.Sprintf("%v ", val))
		}
	}
	return strings.TrimSpace(sb.String())
}

// csvFormatter formats row data into a CSV string.
func csvFormatter(row map[string]any, contentColumns []string) string {
	var sb strings.Builder
	writer := csv.NewWriter(&sb)
	record := make([]string, 0, len(contentColumns))
	for _, column := range contentColumns {
		if val, ok := row[column]; ok {
			record = append(record, fmt.Sprintf("%v", val))
		}
	}
	if err := writer.Write(record); err != nil {
		// Should not happen in normal cases as values are usually simple types
		return ""
	}
	writer.Flush()
	return strings.TrimSuffix(sb.String(), "\n") // Remove trailing newline
}

// yamlFormatter formats row data into a YAML string.
func yamlFormatter(row map[string]any, contentColumns []string) string {
	var sb strings.Builder
	for _, column := range contentColumns {
		if val, ok := row[column]; ok {
			sb.WriteString(fmt.Sprintf("%s: %v\n", column, val))
		}
	}
	return strings.TrimSpace(sb.String())
}

// jsonFormatter formats row data into a JSON string.
func jsonFormatter(row map[string]any, contentColumns []string) string {
	data := make(map[string]any)
	for _, column := range contentColumns {
		if val, ok := row[column]; ok {
			data[column] = val
		}
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		// Should not happen in normal cases as values are usually simple types
		return ""
	}
	return string(jsonData)
}

// parseDocFromRow parses a Document from a row of data.
func (l *DocumentLoader) parseDocFromRow(row map[string]any) schema.Document {
	pageContent := l.formatter(row, l.contentColumns)
	metadata := make(map[string]any)

	if l.metadataJSONColumn != "" {
		value, ok := row[l.metadataJSONColumn]
		if ok {
			mapValues := value.(map[string]any)
			for k, v := range mapValues {
				metadata[k] = v
			}
		}
	}

	for _, column := range l.metadataColumns {
		if column != l.metadataJSONColumn {
			metadata[column] = row[column]
		}
	}

	return schema.Document{
		PageContent: pageContent,
		Metadata:    metadata,
	}
}

// Load executes the configured SQL query and returns a list of Document objects.
func (l *DocumentLoader) Load(ctx context.Context) ([]schema.Document, error) {
	rows, err := l.engine.Pool.Query(ctx, l.query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	fieldDescriptions := rows.FieldDescriptions()
	var documents []schema.Document

	for rows.Next() {
		v, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("unable to parse row: %w", err)
		}
		mapColumnNameValue := make(map[string]any)
		for i, f := range fieldDescriptions {
			mapColumnNameValue[f.Name] = v[i]
		}
		documents = append(documents, l.parseDocFromRow(mapColumnNameValue))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return documents, nil
}

func (l *DocumentLoader) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	if splitter == nil {
		splitter = textsplitter.NewRecursiveCharacter()
	}

	docs, err := l.Load(ctx)
	if err != nil {
		return nil, err
	}
	return textsplitter.SplitDocuments(splitter, docs)
}

func (l *DocumentLoader) validateColumns(columnNames []string) error {
	allNames := make(map[string]struct{})
	for _, name := range l.contentColumns {
		allNames[name] = struct{}{}
	}
	for _, name := range l.metadataColumns {
		allNames[name] = struct{}{}
	}

	for name := range allNames {
		if found := slices.Contains(columnNames, name); found {
			continue
		}
		return fmt.Errorf("column '%s' not found in query result", name)
	}

	return nil
}

func (l *DocumentLoader) getColumNames(ctx context.Context) ([]string, error) {
	rows, err := l.engine.Pool.Query(ctx, fmt.Sprintf("%s LIMIT 1", l.query))
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	colNames := make([]string, len(rows.FieldDescriptions()))
	for i, description := range rows.FieldDescriptions() {
		colNames[i] = description.Name
	}
	return colNames, nil
}

func (l *DocumentLoader) configureColumns(columnNames []string) error {
	if len(l.contentColumns) == 0 {
		l.contentColumns = []string{columnNames[0]}
	}

	if len(l.metadataColumns) == 0 {
		for _, col := range columnNames {
			if !slices.Contains(l.contentColumns, col) {
				l.metadataColumns = append(l.metadataColumns, col)
			}
		}
	}

	if strings.TrimSpace(l.metadataJSONColumn) != "" && !slices.Contains(columnNames, l.metadataJSONColumn) {
		return fmt.Errorf("column '%s' not found in query result %v", l.metadataJSONColumn, columnNames)
	}

	if l.metadataJSONColumn == "" && slices.Contains(columnNames, defaultMetadataJSONColumn) {
		l.metadataJSONColumn = defaultMetadataJSONColumn
	}

	return nil
}
