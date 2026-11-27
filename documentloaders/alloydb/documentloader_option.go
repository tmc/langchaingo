package alloydb

import (
	"fmt"
	"regexp"
)

const sqlregularexpresion = `(?i)^\s*SELECT\s+.+\s+FROM\s+((")?([a-zA-Z0-9_]+)(")?\.)?(")?([a-zA-Z0-9_]+)(")?\b`

// DocumentLoaderOption is a functional option for configuring the DocumentLoader.
type DocumentLoaderOption func(*DocumentLoader)

// WithSchemaName sets the schema name for the table. Defaults to "public".
func WithSchemaName(schemaName string) DocumentLoaderOption {
	return func(documentLoader *DocumentLoader) {
		documentLoader.schemaName = schemaName
	}
}

// WithQuery sets the SQL query to execute. If not provided, a default query is generated from the table name.
func WithQuery(query string) DocumentLoaderOption {
	return func(documentLoader *DocumentLoader) {
		documentLoader.query = query
	}
}

// WithTableName sets the table name to load data from. If not provided, a custom query must be specified.
func WithTableName(tableName string) DocumentLoaderOption {
	return func(documentLoader *DocumentLoader) {
		documentLoader.tableName = tableName
	}
}

// WithCustomFormatter sets a custom formatter to convert row data into document content.
func WithCustomFormatter(formatter func(map[string]interface{}, []string) string) DocumentLoaderOption {
	return func(documentLoader *DocumentLoader) {
		documentLoader.formatter = formatter
	}
}

// WithJSONFormatter sets a json formatter to convert row data into document content.
func WithJSONFormatter() DocumentLoaderOption {
	return func(documentLoader *DocumentLoader) {
		documentLoader.formatter = jsonFormatter
	}
}

// WithTextFormatter sets a text formatter to convert row data into document content.
func WithTextFormatter() DocumentLoaderOption {
	return func(documentLoader *DocumentLoader) {
		documentLoader.formatter = textFormatter
	}
}

// WithYAMLFormatter sets a yaml formatter to convert row data into document content.
func WithYAMLFormatter() DocumentLoaderOption {
	return func(documentLoader *DocumentLoader) {
		documentLoader.formatter = yamlFormatter
	}
}

// WithCSVFormatter sets a csv formatter to convert row data into document content.
func WithCSVFormatter() DocumentLoaderOption {
	return func(documentLoader *DocumentLoader) {
		documentLoader.formatter = csvFormatter
	}
}

// WithContentColumns sets the list of columns to use for the document content.
func WithContentColumns(contentColumns []string) DocumentLoaderOption {
	return func(documentLoader *DocumentLoader) {
		documentLoader.contentColumns = contentColumns
	}
}

// WithMetadataColumns sets the list of columns to use for the document metadata.
func WithMetadataColumns(metadataColumns []string) DocumentLoaderOption {
	return func(documentLoader *DocumentLoader) {
		documentLoader.metadataColumns = metadataColumns
	}
}

// WithMetadataJSONColumn sets the column name containing JSON metadata.
func WithMetadataJSONColumn(metadataJSONColumn string) DocumentLoaderOption {
	return func(documentLoader *DocumentLoader) {
		documentLoader.metadataJSONColumn = metadataJSONColumn
	}
}

func validateDocumentLoader(dl *DocumentLoader) error {
	if dl.engine.Pool == nil {
		return fmt.Errorf("engine.Pool must be specified")
	}

	if dl.query == "" && dl.tableName == "" {
		return fmt.Errorf("either query or tableName must be specified")
	}

	if dl.query != "" && dl.tableName != "" {
		return fmt.Errorf("only one of 'table_name' or 'query' should be specified")
	}

	if dl.query == "" {
		dl.query = fmt.Sprintf(`SELECT * FROM "%s"."%s"`, dl.schemaName, dl.tableName)
	}

	if dl.formatter == nil {
		// default formatter
		dl.formatter = textFormatter
	}
	return nil
}

func validateQuery(query string) error {
	re := regexp.MustCompile(sqlregularexpresion)
	if !re.MatchString(query) {
		return fmt.Errorf("query is not valid for the following regular expression: %s", sqlregularexpresion)
	}
	return nil
}
