package alloydb

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"cloud.google.com/go/alloydbconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tmc/langchaingo/vectorstores/pgvector"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

// ServiceAccountRetriever defines an interface to determine the IAM account
// email when user authentication is not explicitly provided.
type serviceAccountRetriever interface {
	serviceAccountEmailGetter(ctx context.Context) (string, error)
}
type DefaultServiceAccountRetriever struct{}

func serviceAccountEmailGetter(ctx context.Context) (string, error) {
	return getServiceAccountEmail(ctx)
}

type Column struct {
	Name     string
	DataType string
	Nullable bool
}

type PostgresEngineConfig struct {
	conn                    pgvector.PGXConn
	projectId               string
	region                  string
	cluster                 string
	instance                string
	host                    string
	port                    int
	database                string
	user                    string
	password                string
	ipType                  string
	iAmAccountEmail         string
	serviceAccountRetriever serviceAccountRetriever
}

// NewPostgresEngineConfig initializes a new PostgresEngineConfig
func NewPostgresEngineConfig(ctx context.Context, opts ...Option) (pgEngineConfig *PostgresEngineConfig, err error) {
	pgEngineConfig = applyClientOptions(opts...)
	// TODO:: Handle errors and validate mandatory parameters.
	connPool, err := pgEngineConfig.CreateConnection(ctx)
	if err != nil {
		return &PostgresEngineConfig{}, err
	}
	if pgEngineConfig.conn == nil {
		pgEngineConfig.conn = connPool
	}
	return pgEngineConfig, nil
}

// CreateConnection creates a connection pool to the PostgreSQL database.
func (p *PostgresEngineConfig) CreateConnection(ctx context.Context) (*pgxpool.Pool, error) {
	username, _, err := p.assignUser(ctx) // TODO :: usingIAMAuth >> add oauth2 google credentials auth
	if err != nil {
		return nil, fmt.Errorf("error assigning user. Err: %w", err)
	}

	// Configure the driver to connect to the database
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", username, p.password, p.database)
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection config: %w", err)
	}

	// Create a new dialer with any options
	d, err := alloydbconn.NewDialer(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize connection: %w", err)
	}
	// Don't close the dialer until you're done with the database connection
	defer d.Close()

	// Create the connection
	config.ConnConfig.DialFunc = func(ctx context.Context, _ string, instance string) (net.Conn, error) {
		return d.Dial(ctx, "projects/<PROJECT>/locations/<REGION>/clusters/<CLUSTER>/instances/<INSTANCE>")
	}

	// TODO :: If only async connection will be used, here we won't connect, only a connection will be returned.
	// Will delete this later, only for testing purposes

	conn, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}
	return conn, nil
}

func (p *PostgresEngineConfig) assignUser(ctx context.Context) (username string, usingIAMAuth bool, err error) {
	// If neither user nor password are provided, retrieve IAM email
	if p.user == "" && p.password == "" {
		username, err := p.serviceAccountRetriever.serviceAccountEmailGetter(ctx)
		if err != nil {
			return "", false, fmt.Errorf("unable to retrieve service account email: %w", err)
		}
		return username, true, nil
	} else if p.user != "" && p.password != "" {
		// If both username and password are provided use default username
		return p.user, false, nil
	}

	// If no user can be determined, return an error.
	return "", false, errors.New("unable to retrieve a valid username")
}

// getServiceAccountEmail retrieves the IAM principal email with users account.
func getServiceAccountEmail(ctx context.Context) (string, error) {
	scopes := []string{"https://www.googleapis.com/auth/userinfo.email"}
	// Get credentials using email scope
	credentials, err := google.FindDefaultCredentials(ctx, scopes...) // TODO :: Additional scopes will be added in the Cloud SQL Go Connector.
	if err != nil {
		return "", fmt.Errorf("unable to get default credentials: %s", err)
	}

	// Verify valid TokenSource
	if credentials.TokenSource == nil {
		return "", fmt.Errorf("missing or invalid credentials")
	}

	oauth2Service, err := oauth2.NewService(ctx, option.WithTokenSource(credentials.TokenSource))
	if err != nil {
		return "", fmt.Errorf("failed to create new service: %v", err)
	}

	// Fetch IAM principal email
	userInfo, err := oauth2Service.Userinfo.Get().Do()
	if err != nil {
		return "", fmt.Errorf("failed to get user info: %v", err)
	}
	return userInfo.Email, nil
}

// getConnAsync generates an asynchronous connection.
func (p *PostgresEngineConfig) getConnAsync(ctx context.Context) (*pgxpool.Pool, error) {
	connCh := make(chan *pgxpool.Pool, 1)
	errCh := make(chan error, 1)

	go func() {
		conn, err := p.CreateConnection(ctx)
		if err != nil {
			errCh <- err
			close(errCh)
			return
		}
		connCh <- conn
		close(connCh)
	}()

	// Wait for the result or error
	select {
	case conn := <-connCh:
		if conn == nil {
			return nil, fmt.Errorf("failed to establish connection")
		}
		return conn, nil
	case err := <-errCh:
		return nil, fmt.Errorf("unable to connect: %w", err)
	}
}

// startBackgroundLoop starts a new goroutine that runs the background.
func (p *PostgresEngineConfig) startBackgroundLoop(ctx context.Context, wg *sync.WaitGroup, resultChan chan<- *pgxpool.Pool, errChan chan<- error) {
	// Notify that the goroutine is done when it finishes
	defer wg.Done()

	// Create the PostgresEngine asynchronously in a goroutine
	go func() {
		// Establish the connection
		conn, err := p.CreateConnection(ctx)
		if err != nil {
			log.Printf("Error creating the PostgresE ngine: %v", err)
			errChan <- err
			resultChan <- nil
			return
		}

		// Return the connection via the channel
		resultChan <- conn
	}()
}

// TODO :: conn methods should be called with interfaces?
// TODO :: here I should have another options to have the multiple and default values of the parameters.
//
//	initVectorstoreTable creates a table for saving of vectors to be used with PostgresVectorStore.
func (p *PostgresEngineConfig) initVectorstoreTable(ctx context.Context, tableName string, vectorSize int, schemaName string, contentColumn string,
	embeddingColumn string, metadataColumns []Column, metadataJsonColumn string, idColumn interface{}, overwriteExisting bool, storeMetadata bool) error {
	// Ensure the vector extension exists
	_, err := p.conn.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		return fmt.Errorf("failed to create extension: %v", err)
	}

	// Drop table if exists and overwrite flag is true
	if overwriteExisting {
		_, err = p.conn.Exec(ctx, fmt.Sprintf(`DROP TABLE IF EXISTS "%s"."%s"`, schemaName, tableName)) // SchemaName default is public.
		if err != nil {
			return fmt.Errorf("failed to drop table: %v", err)
		}
	}

	// Determine the id column type and name
	var idDataType, idColumnName string
	if idStr, ok := idColumn.(string); ok {
		idDataType = "UUID"
		idColumnName = idStr
	} else if col, ok := idColumn.(Column); ok {
		idDataType = col.DataType
		idColumnName = col.Name
	}

	// Build the SQL query that creates the table
	query := fmt.Sprintf(`CREATE TABLE "%s"."%s" (
		"%s" %s PRIMARY KEY,
		"%s" TEXT NOT NULL,
		"%s" vector(%d) NOT NULL`, schemaName, tableName, idColumnName, idDataType, contentColumn, embeddingColumn, vectorSize)

	// Add metadata columns  to the query string if provided
	for _, column := range metadataColumns {
		nullable := ""
		if !column.Nullable {
			nullable = "NOT NULL"
		}
		query += fmt.Sprintf(`, "%s" %s %s`, column.Name, column.DataType, nullable)
	}

	// Add JSON metadata column to the query string if storeMetadata is true
	if storeMetadata {
		query += fmt.Sprintf(`, "%s" JSON`, metadataJsonColumn)
	}
	// Close the query string
	query += ");"

	// TODO :: this query must be asynchronous

	// Execute the query to create the table
	_, err = p.conn.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	return nil
}

// initChatHistoryTable creates a Cloud SQL table to store chat history.
func (p *PostgresEngineConfig) initChatHistoryTable(ctx context.Context, tableName string, schemaName string) error {
	createTableQuery := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s"."%s" (
		id SERIAL PRIMARY KEY,
		session_id TEXT NOT NULL,
		data JSONB NOT NULL,
		type TEXT NOT NULL
	);`, schemaName, tableName)

	// Execute the query
	_, err := p.conn.Exec(ctx, createTableQuery)
	if err != nil {
		return fmt.Errorf("failed to execute query: %v", err)
	}
	return nil
}
