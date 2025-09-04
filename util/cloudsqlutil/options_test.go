package cloudsqlutil

import (
	"testing"
)

// Unit tests that don't require external dependencies

func TestApplyClientOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
		check   func(*testing.T, engineConfig)
	}{
		{
			name: "default configuration",
			opts: []Option{
				WithCloudSQLInstance("project", "region", "instance"),
			},
			wantErr: false,
			check: func(t *testing.T, cfg engineConfig) {
				if cfg.projectID != "project" {
					t.Errorf("projectID = %q, want %q", cfg.projectID, "project")
				}
				if cfg.region != "region" {
					t.Errorf("region = %q, want %q", cfg.region, "region")
				}
				if cfg.instance != "instance" {
					t.Errorf("instance = %q, want %q", cfg.instance, "instance")
				}
				if cfg.ipType != "PUBLIC" {
					t.Errorf("ipType = %q, want %q", cfg.ipType, "PUBLIC")
				}
				if cfg.userAgents != defaultUserAgent {
					t.Errorf("userAgents = %q, want %q", cfg.userAgents, defaultUserAgent)
				}
			},
		},
		{
			name: "with all options",
			opts: []Option{
				WithCloudSQLInstance("project", "region", "instance"),
				WithDatabase("testdb"),
				WithUser("testuser"),
				WithPassword("testpass"),
				WithIPType("PRIVATE"),
				WithIAMAccountEmail("test@example.com"),
			},
			wantErr: false,
			check: func(t *testing.T, cfg engineConfig) {
				if cfg.database != "testdb" {
					t.Errorf("database = %q, want %q", cfg.database, "testdb")
				}
				if cfg.user != "testuser" {
					t.Errorf("user = %q, want %q", cfg.user, "testuser")
				}
				if cfg.password != "testpass" {
					t.Errorf("password = %q, want %q", cfg.password, "testpass")
				}
				if cfg.ipType != "PRIVATE" {
					t.Errorf("ipType = %q, want %q", cfg.ipType, "PRIVATE")
				}
				if cfg.iamAccountEmail != "test@example.com" {
					t.Errorf("iamAccountEmail = %q, want %q", cfg.iamAccountEmail, "test@example.com")
				}
			},
		},
		{
			name:    "missing connection fields",
			opts:    []Option{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := applyClientOptions(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyClientOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

func TestValidateVectorstoreTableOptions(t *testing.T) { //nolint:funlen // comprehensive test
	tests := []struct {
		name    string
		opts    VectorstoreTableOptions
		wantErr bool
		check   func(*testing.T, VectorstoreTableOptions)
	}{
		{
			name: "valid minimal options",
			opts: VectorstoreTableOptions{
				TableName:  "test_table",
				VectorSize: 384,
			},
			wantErr: false,
			check: func(t *testing.T, opts VectorstoreTableOptions) {
				if opts.SchemaName != "public" {
					t.Errorf("SchemaName = %q, want %q", opts.SchemaName, "public")
				}
				if opts.ContentColumnName != "content" {
					t.Errorf("ContentColumnName = %q, want %q", opts.ContentColumnName, "content")
				}
				if opts.EmbeddingColumn != "embedding" {
					t.Errorf("EmbeddingColumn = %q, want %q", opts.EmbeddingColumn, "embedding")
				}
				if opts.MetadataJSONColumn != "langchain_metadata" {
					t.Errorf("MetadataJSONColumn = %q, want %q", opts.MetadataJSONColumn, "langchain_metadata")
				}
				if opts.IDColumn.Name != "langchain_id" {
					t.Errorf("IDColumn.Name = %q, want %q", opts.IDColumn.Name, "langchain_id")
				}
				if opts.IDColumn.DataType != "UUID" {
					t.Errorf("IDColumn.DataType = %q, want %q", opts.IDColumn.DataType, "UUID")
				}
			},
		},
		{
			name: "valid options with custom values",
			opts: VectorstoreTableOptions{
				TableName:          "custom_table",
				VectorSize:         768,
				SchemaName:         "custom_schema",
				ContentColumnName:  "custom_content",
				EmbeddingColumn:    "custom_embedding",
				MetadataJSONColumn: "custom_metadata",
				IDColumn: Column{
					Name:     "custom_id",
					DataType: "SERIAL",
				},
			},
			wantErr: false,
			check: func(t *testing.T, opts VectorstoreTableOptions) {
				if opts.SchemaName != "custom_schema" {
					t.Errorf("SchemaName = %q, want %q", opts.SchemaName, "custom_schema")
				}
				if opts.ContentColumnName != "custom_content" {
					t.Errorf("ContentColumnName = %q, want %q", opts.ContentColumnName, "custom_content")
				}
				if opts.EmbeddingColumn != "custom_embedding" {
					t.Errorf("EmbeddingColumn = %q, want %q", opts.EmbeddingColumn, "custom_embedding")
				}
				if opts.MetadataJSONColumn != "custom_metadata" {
					t.Errorf("MetadataJSONColumn = %q, want %q", opts.MetadataJSONColumn, "custom_metadata")
				}
				if opts.IDColumn.Name != "custom_id" {
					t.Errorf("IDColumn.Name = %q, want %q", opts.IDColumn.Name, "custom_id")
				}
				if opts.IDColumn.DataType != "SERIAL" {
					t.Errorf("IDColumn.DataType = %q, want %q", opts.IDColumn.DataType, "SERIAL")
				}
			},
		},
		{
			name: "missing table name",
			opts: VectorstoreTableOptions{
				VectorSize: 384,
			},
			wantErr: true,
		},
		{
			name: "missing vector size",
			opts: VectorstoreTableOptions{
				TableName: "test_table",
			},
			wantErr: true,
		},
		{
			name: "zero vector size",
			opts: VectorstoreTableOptions{
				TableName:  "test_table",
				VectorSize: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalOpts := tt.opts
			err := validateVectorstoreTableOptions(&tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateVectorstoreTableOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, tt.opts)
			}
			// Verify that the original struct was modified
			if !tt.wantErr && originalOpts.SchemaName == "" {
				if tt.opts.SchemaName != "public" {
					t.Error("opts should have been modified with default values")
				}
			}
		})
	}
}

func TestApplyChatMessageHistoryOptions(t *testing.T) {
	tests := []struct {
		name string
		opts []OptionInitChatHistoryTable
		want InitChatHistoryTableOptions
	}{
		{
			name: "default options",
			opts: []OptionInitChatHistoryTable{},
			want: InitChatHistoryTableOptions{
				schemaName: "public",
			},
		},
		{
			name: "custom schema name",
			opts: []OptionInitChatHistoryTable{
				WithSchemaName("custom_schema"),
			},
			want: InitChatHistoryTableOptions{
				schemaName: "custom_schema",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyChatMessageHistoryOptions(tt.opts...)
			if got.schemaName != tt.want.schemaName {
				t.Errorf("applyChatMessageHistoryOptions() schemaName = %v, want %v", got.schemaName, tt.want.schemaName)
			}
		})
	}
}

func TestOptions(t *testing.T) {
	tests := []struct {
		name   string
		option Option
		check  func(*testing.T, engineConfig)
	}{
		{
			name:   "WithCloudSQLInstance",
			option: WithCloudSQLInstance("project1", "region1", "instance1"),
			check: func(t *testing.T, cfg engineConfig) {
				if cfg.projectID != "project1" {
					t.Errorf("projectID = %q, want %q", cfg.projectID, "project1")
				}
				if cfg.region != "region1" {
					t.Errorf("region = %q, want %q", cfg.region, "region1")
				}
				if cfg.instance != "instance1" {
					t.Errorf("instance = %q, want %q", cfg.instance, "instance1")
				}
			},
		},
		{
			name:   "WithDatabase",
			option: WithDatabase("testdb"),
			check: func(t *testing.T, cfg engineConfig) {
				if cfg.database != "testdb" {
					t.Errorf("database = %q, want %q", cfg.database, "testdb")
				}
			},
		},
		{
			name:   "WithUser",
			option: WithUser("testuser"),
			check: func(t *testing.T, cfg engineConfig) {
				if cfg.user != "testuser" {
					t.Errorf("user = %q, want %q", cfg.user, "testuser")
				}
			},
		},
		{
			name:   "WithPassword",
			option: WithPassword("testpass"),
			check: func(t *testing.T, cfg engineConfig) {
				if cfg.password != "testpass" {
					t.Errorf("password = %q, want %q", cfg.password, "testpass")
				}
			},
		},
		{
			name:   "WithIPType",
			option: WithIPType("PRIVATE"),
			check: func(t *testing.T, cfg engineConfig) {
				if cfg.ipType != "PRIVATE" {
					t.Errorf("ipType = %q, want %q", cfg.ipType, "PRIVATE")
				}
			},
		},
		{
			name:   "WithIAMAccountEmail",
			option: WithIAMAccountEmail("test@example.com"),
			check: func(t *testing.T, cfg engineConfig) {
				if cfg.iamAccountEmail != "test@example.com" {
					t.Errorf("iamAccountEmail = %q, want %q", cfg.iamAccountEmail, "test@example.com")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &engineConfig{}
			tt.option(cfg)
			tt.check(t, *cfg)
		})
	}
}

func TestChatHistoryTableOptions(t *testing.T) {
	t.Run("WithSchemaName", func(t *testing.T) {
		opts := &InitChatHistoryTableOptions{}
		WithSchemaName("test_schema")(opts)
		if opts.schemaName != "test_schema" {
			t.Errorf("schemaName = %q, want %q", opts.schemaName, "test_schema")
		}
	})
}

func TestColumnStruct(t *testing.T) {
	column := Column{
		Name:     "test_column",
		DataType: "VARCHAR(255)",
		Nullable: true,
	}

	if column.Name != "test_column" {
		t.Errorf("Name = %q, want %q", column.Name, "test_column")
	}
	if column.DataType != "VARCHAR(255)" {
		t.Errorf("DataType = %q, want %q", column.DataType, "VARCHAR(255)")
	}
	if !column.Nullable {
		t.Error("Nullable should be true")
	}
}

func TestVectorstoreTableOptionsStruct(t *testing.T) {
	opts := VectorstoreTableOptions{
		TableName:          "test_table",
		VectorSize:         768,
		SchemaName:         "test_schema",
		ContentColumnName:  "content_col",
		EmbeddingColumn:    "embed_col",
		MetadataJSONColumn: "meta_col",
		IDColumn: Column{
			Name:     "id_col",
			DataType: "UUID",
			Nullable: false,
		},
		MetadataColumns: []Column{
			{Name: "title", DataType: "TEXT", Nullable: true},
			{Name: "category", DataType: "VARCHAR(100)", Nullable: false},
		},
		OverwriteExisting: true,
		StoreMetadata:     true,
	}

	if opts.TableName != "test_table" {
		t.Errorf("TableName = %q, want %q", opts.TableName, "test_table")
	}
	if opts.VectorSize != 768 {
		t.Errorf("VectorSize = %d, want %d", opts.VectorSize, 768)
	}
	if opts.SchemaName != "test_schema" {
		t.Errorf("SchemaName = %q, want %q", opts.SchemaName, "test_schema")
	}
	if opts.ContentColumnName != "content_col" {
		t.Errorf("ContentColumnName = %q, want %q", opts.ContentColumnName, "content_col")
	}
	if opts.EmbeddingColumn != "embed_col" {
		t.Errorf("EmbeddingColumn = %q, want %q", opts.EmbeddingColumn, "embed_col")
	}
	if opts.MetadataJSONColumn != "meta_col" {
		t.Errorf("MetadataJSONColumn = %q, want %q", opts.MetadataJSONColumn, "meta_col")
	}
	if opts.IDColumn.Name != "id_col" {
		t.Errorf("IDColumn.Name = %q, want %q", opts.IDColumn.Name, "id_col")
	}
	if opts.IDColumn.DataType != "UUID" {
		t.Errorf("IDColumn.DataType = %q, want %q", opts.IDColumn.DataType, "UUID")
	}
	if opts.IDColumn.Nullable {
		t.Error("IDColumn.Nullable should be false")
	}
	if len(opts.MetadataColumns) != 2 {
		t.Errorf("MetadataColumns length = %d, want %d", len(opts.MetadataColumns), 2)
	}
	if !opts.OverwriteExisting {
		t.Error("OverwriteExisting should be true")
	}
	if !opts.StoreMetadata {
		t.Error("StoreMetadata should be true")
	}
}

func TestConstants(t *testing.T) {
	if defaultSchemaName != "public" {
		t.Errorf("defaultSchemaName = %q, want %q", defaultSchemaName, "public")
	}
	if defaultUserAgent != "langchaingo-cloud-sql-pg/0.0.0" {
		t.Errorf("defaultUserAgent = %q, want %q", defaultUserAgent, "langchaingo-cloud-sql-pg/0.0.0")
	}
}

func TestPostgresEngineClose(t *testing.T) {
	// Test closing with nil pool
	engine := &PostgresEngine{Pool: nil}
	engine.Close() // Should not panic

	// Note: Testing with actual pool would require integration test setup
}
