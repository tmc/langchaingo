package cloudsql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/util/cloudsqlutil"
)

func TestDocumentLoaderOption(t *testing.T) {
	t.Parallel()
	testEngine, teardown, err := setup(t)
	require.NoError(t, err)
	t.Cleanup(teardown)
	type args struct {
		engine  cloudsqlutil.PostgresEngine
		options []DocumentLoaderOption
	}
	tests := []struct {
		name         string
		args         args
		wantErr      bool
		validateFunc func(t *testing.T, c *DocumentLoader, err error)
	}{
		{
			name:         "pgxpool is nil",
			args:         args{engine: cloudsqlutil.PostgresEngine{}},
			wantErr:      true,
			validateFunc: validatePgxIsNil(),
		},
		{
			name:         "table name and query are empty",
			args:         args{engine: testEngine},
			wantErr:      true,
			validateFunc: validateTableNameAndQueryAreEmpty(),
		},
		{
			name: "table name and query are not empty",
			args: args{
				engine:  testEngine,
				options: []DocumentLoaderOption{WithTableName("testtable"), WithQuery("select * from public.test")},
			},
			wantErr:      true,
			validateFunc: validateTableNameAndQueryAreNotEmpty(),
		},
		{
			name: "success",
			args: args{
				engine:  testEngine,
				options: []DocumentLoaderOption{WithTableName("testtable"), WithJSONFormatter()},
			},
			wantErr:      false,
			validateFunc: validateSuccess(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := &DocumentLoader{engine: tt.args.engine, schemaName: defaultSchemaName}
			for _, opt := range tt.args.options {
				opt(dl)
			}

			err := validateDocumentLoader(dl)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDocumentLoader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.validateFunc(t, dl, err)
		})
	}
}

/************************
	HELPER FUNCTIONS
*************************/

func validateSuccess() func(t *testing.T, c *DocumentLoader, err error) {
	return func(t *testing.T, c *DocumentLoader, err error) {
		t.Helper()
		assert.Nil(t, err)
		require.NotNil(t, c)
		assert.Equal(t, c.query, "SELECT * FROM \"public\".\"testtable\"")
	}
}

func validateTableNameAndQueryAreNotEmpty() func(t *testing.T, c *DocumentLoader, err error) {
	return func(t *testing.T, c *DocumentLoader, err error) {
		t.Helper()
		assert.NotNil(t, c)
		assert.EqualError(t, err, "only one of 'table_name' or 'query' should be specified")
	}
}

func validateTableNameAndQueryAreEmpty() func(t *testing.T, c *DocumentLoader, err error) {
	return func(t *testing.T, c *DocumentLoader, err error) {
		t.Helper()
		assert.NotNil(t, c)
		assert.EqualError(t, err, "either query or tableName must be specified")
	}
}

func validatePgxIsNil() func(t *testing.T, c *DocumentLoader, err error) {
	return func(t *testing.T, c *DocumentLoader, err error) {
		t.Helper()
		assert.NotNil(t, c)
		assert.EqualError(t, err, "engine.Pool must be specified")
	}
}
