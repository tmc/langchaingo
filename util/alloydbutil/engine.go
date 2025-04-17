package alloydbutil

import (
	"context"
	"errors"
	"fmt"
	"net"

	"cloud.google.com/go/alloydbconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

type EmailRetriever func(context.Context) (string, error)

type PostgresEngine struct {
	Pool *pgxpool.Pool
}

type Column struct {
	Name     string
	DataType string
	Nullable bool
}

// NewPostgresEngine creates a new PostgresEngine.
func NewPostgresEngine(ctx context.Context, opts ...Option) (PostgresEngine, error) {
	pgEngine := new(PostgresEngine)
	cfg, err := applyClientOptions(opts...)
	if err != nil {
		return PostgresEngine{}, err
	}
	if cfg.connPool == nil {
		user, usingIAMAuth, err := getUser(ctx, cfg)
		if err != nil {
			return PostgresEngine{}, fmt.Errorf("error assigning user. Err: %w", err)
		}
		if usingIAMAuth {
			cfg.user = user
		}
		cfg.connPool, err = createPool(ctx, cfg, usingIAMAuth)
		if err != nil {
			return PostgresEngine{}, err
		}
	}
	pgEngine.Pool = cfg.connPool
	return *pgEngine, nil
}

// createPool creates a connection pool to the PostgreSQL database.
func createPool(ctx context.Context, cfg engineConfig, usingIAMAuth bool) (*pgxpool.Pool, error) {
	dialeropts := []alloydbconn.Option{alloydbconn.WithUserAgent(cfg.userAgents)}
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", cfg.user, cfg.password, cfg.database)
	if usingIAMAuth {
		dialeropts = append(dialeropts, alloydbconn.WithIAMAuthN())
		dsn = fmt.Sprintf("user=%s dbname=%s sslmode=disable", cfg.user, cfg.database)
	}
	d, err := alloydbconn.NewDialer(ctx, dialeropts...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize connection: %w", err)
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection config: %w", err)
	}
	instanceURI := fmt.Sprintf("projects/%s/locations/%s/clusters/%s/instances/%s", cfg.projectID, cfg.region, cfg.cluster, cfg.instance)
	config.ConnConfig.DialFunc = func(ctx context.Context, _ string, _ string) (net.Conn, error) {
		if cfg.ipType == "PRIVATE" {
			return d.Dial(ctx, instanceURI, alloydbconn.WithPrivateIP())
		}
		return d.Dial(ctx, instanceURI, alloydbconn.WithPublicIP())
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}
	return pool, nil
}

// Close closes the connection.
func (p *PostgresEngine) Close() {
	if p.Pool != nil {
		// Close the connection pool.
		p.Pool.Close()
	}
}

// getUser retrieves the username, a flag indicating if IAM authentication
// will be used and an error.
func getUser(ctx context.Context, config engineConfig) (string, bool, error) {
	switch {
	case config.user != "" && config.password != "":
		// If both username and password are provided use provided username.
		return config.user, false, nil
	case config.iamAccountEmail != "":
		// If iamAccountEmail is provided use it as user.
		return config.iamAccountEmail, true, nil
	case config.user == "" && config.password == "" && config.iamAccountEmail == "":
		// If neither user and password nor iamAccountEmail are provided,
		// retrieve IAM email from the environment.
		serviceAccountEmail, err := config.emailRetriever(ctx)
		if err != nil {
			return "", false, fmt.Errorf("unable to retrieve service account email: %w", err)
		}
		return serviceAccountEmail, true, nil
	}

	// If no user can be determined, return an error.
	return "", false, errors.New("unable to retrieve a valid username")
}

// getServiceAccountEmail retrieves the IAM principal email with users account.
func getServiceAccountEmail(ctx context.Context) (string, error) {
	scopes := []string{"https://www.googleapis.com/auth/userinfo.email"}
	// Get credentials using email scope
	credentials, err := google.FindDefaultCredentials(ctx, scopes...)
	if err != nil {
		return "", fmt.Errorf("unable to get default credentials: %w", err)
	}

	// Verify valid TokenSource.
	if credentials.TokenSource == nil {
		return "", fmt.Errorf("missing or invalid credentials")
	}

	oauth2Service, err := oauth2.NewService(ctx, option.WithTokenSource(credentials.TokenSource))
	if err != nil {
		return "", fmt.Errorf("failed to create new service: %w", err)
	}

	// Fetch IAM principal email.
	userInfo, err := oauth2Service.Userinfo.Get().Do()
	if err != nil {
		return "", fmt.Errorf("failed to get user info: %w", err)
	}
	return userInfo.Email, nil
}

// initChatHistoryTable creates a table to store chat history.
func (p *PostgresEngine) InitChatHistoryTable(ctx context.Context, tableName string, opts ...OptionInitChatHistoryTable) error {
	cfg := applyChatMessageHistoryOptions(opts...)

	createTableQuery := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s"."%s" (
		id SERIAL PRIMARY KEY,
		session_id TEXT NOT NULL,
		data JSONB NOT NULL,
		type TEXT NOT NULL
	);`, cfg.schemaName, tableName)

	// Execute the query
	_, err := p.Pool.Exec(ctx, createTableQuery)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

// validateVectorstoreTableOptions initializes the options struct with the default values for
// the InitVectorstoreTable function.
func validateVectorstoreTableOptions(opts *VectorstoreTableOptions) error {
	if opts.TableName == "" {
		return fmt.Errorf("missing table name in options")
	}
	if opts.VectorSize == 0 {
		return fmt.Errorf("missing vector size in options")
	}

	if opts.SchemaName == "" {
		opts.SchemaName = "public"
	}

	if opts.ContentColumnName == "" {
		opts.ContentColumnName = "content"
	}

	if opts.EmbeddingColumn == "" {
		opts.EmbeddingColumn = "embedding"
	}

	if opts.MetadataJSONColumn == "" {
		opts.MetadataJSONColumn = "langchain_metadata"
	}

	if opts.IDColumn.Name == "" {
		opts.IDColumn.Name = "langchain_id"
	}

	if opts.IDColumn.DataType == "" {
		opts.IDColumn.DataType = "UUID"
	}

	return nil
}

// initVectorstoreTable creates a table for saving of vectors to be used with PostgresVectorStore.
func (p *PostgresEngine) InitVectorstoreTable(ctx context.Context, opts VectorstoreTableOptions) error {
	err := validateVectorstoreTableOptions(&opts)
	if err != nil {
		return fmt.Errorf("failed to validate vectorstore table options: %w", err)
	}

	// Ensure the vector extension exists
	_, err = p.Pool.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		return fmt.Errorf("failed to create extension: %w", err)
	}

	// Drop table if exists and overwrite flag is true
	if opts.OverwriteExisting {
		_, err = p.Pool.Exec(ctx, fmt.Sprintf(`DROP TABLE IF EXISTS "%s"."%s"`, opts.SchemaName, opts.TableName))
		if err != nil {
			return fmt.Errorf("failed to drop table: %w", err)
		}
	}

	// Build the SQL query that creates the table
	query := fmt.Sprintf(`CREATE TABLE "%s"."%s" (
		"%s" %s PRIMARY KEY,
		"%s" TEXT NOT NULL,
		"%s" vector(%d) NOT NULL`, opts.SchemaName, opts.TableName, opts.IDColumn.Name, opts.IDColumn.DataType, opts.ContentColumnName, opts.EmbeddingColumn, opts.VectorSize)

	// Add metadata columns  to the query string if provided
	for _, column := range opts.MetadataColumns {
		nullable := ""
		if !column.Nullable {
			nullable = "NOT NULL"
		}
		query += fmt.Sprintf(`, "%s" %s %s`, column.Name, column.DataType, nullable)
	}

	// Add JSON metadata column to the query string if storeMetadata is true
	if opts.StoreMetadata {
		query += fmt.Sprintf(`, "%s" JSON`, opts.MetadataJSONColumn)
	}
	// Close the query string
	query += ");"

	// Execute the query to create the table
	_, err = p.Pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}
