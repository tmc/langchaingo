package alloydb

import (
	"context"
	"errors"
	"fmt"
	"net"

	"cloud.google.com/go/alloydbconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
	"google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

type EmailRetreiver func(context.Context) (string, error)

type PostgresEngine struct {
	pool *pgxpool.Pool
}

// NewPostgresEngine creates a new PostgresEngine.
func NewPostgresEngine(ctx context.Context, opts ...Option) (*PostgresEngine, error) {
	pgEngine := new(PostgresEngine)
	cfg, err := applyClientOptions(opts...)
	if err != nil {
		return nil, err
	}
	username, usingIAMAuth, err := getUser(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("error assigning user. Err: %w", err)
	}
	if usingIAMAuth {
		token, err := getIAMToken(ctx, username)
		if err != nil {
			return nil, err
		}
		cfg.password = token
	}
	if cfg.connPool == nil {
		cfg.connPool, err = createConnection(ctx, cfg)
		if err != nil {
			return &PostgresEngine{}, err
		}
	}
	pgEngine.pool = cfg.connPool
	return pgEngine, nil
}

// createConnection creates a connection pool to the PostgreSQL database.
func createConnection(ctx context.Context, cfg engineConfig) (*pgxpool.Pool, error) {
	// Create a new dialer with any options
	d, err := alloydbconn.NewDialer(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize connection: %w", err)
	}

	// Configure the driver to connect to the database
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", cfg.user, cfg.password, cfg.database)
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection config: %w", err)
	}

	instanceURI := fmt.Sprintf("projects/%s/locations/%s/clusters/%s/instances/%s", cfg.projectID, cfg.region, cfg.cluster, cfg.instance)
	// Create the connection
	config.ConnConfig.DialFunc = func(ctx context.Context, _ string, _ string) (net.Conn, error) {
		return d.Dial(ctx, instanceURI)
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}
	return pool, nil
}

// Close closes the connection.
func (p *PostgresEngine) Close() {
	if p.pool != nil {
		// Close the connection pool.
		p.pool.Close()
	}
}

func getUser(ctx context.Context, config engineConfig) (username string, usingIAMAuth bool, err error) {
	// If neither user nor password are provided, retrieve IAM email.
	if config.user == "" && config.password == "" {
		serviceAccountEmail, err := config.emailRetreiver(ctx)
		if err != nil {
			return "", false, fmt.Errorf("unable to retrieve service account email: %w", err)
		}
		return serviceAccountEmail, true, nil
	} else if config.user != "" && config.password != "" {
		// If both username and password are provided use default username.
		return config.user, false, nil
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

// getIAMToken retrieves the IAM token for a service account.
func getIAMToken(ctx context.Context, serviceAccountEmail string) (string, error) {
	// Create the OAuth2 token source using the service account's credentials
	tokenSource, err := idtoken.NewTokenSource(ctx, serviceAccountEmail)
	if err != nil {
		return "", fmt.Errorf("failed to create token source: %w", err)
	}

	// Get the IAM token for authentication
	token, err := tokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to get IAM token: %w", err)
	}

	return token.AccessToken, nil
}
