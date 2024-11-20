package pgvector

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

const Port = 5432 // default PostgreSQL port

type ServiceAccountRetriever interface {
	GetServiceAccountEmail(ctx context.Context) (string, error)
}
type DefaultServiceAccountRetriever struct{}

func (d *DefaultServiceAccountRetriever) GetServiceAccountEmail(ctx context.Context) (string, error) {
	return getServiceAccountEmail(ctx)
}

type Column struct {
	Name     string
	DataType string
	Nullable bool
}

func NewColumn(name, dataType string, nullable bool) (*Column, error) {
	if name == "" {
		return nil, errors.New("name should be provided")
	}
	if dataType == "" {
		return nil, errors.New("dataType should be provided")
	}
	return &Column{Name: name, DataType: dataType, Nullable: nullable}, nil
}

type PostgresEngine struct {
	Host                    string
	Port                    int
	Database                string
	User                    string
	Password                string
	IAMAccountEmail         string
	ServiceAccountRetriever ServiceAccountRetriever
}

// NewPostgresEngine initializes a new PostgresEngine
func NewPostgresEngine(host, dbname, user, password, iamAccountEmail string, port int) *PostgresEngine {
	return &PostgresEngine{
		Host:            host,
		Port:            port,
		Database:        dbname,
		User:            user,
		Password:        password,
		IAMAccountEmail: iamAccountEmail,
	}
}

// CreatePostgresEngine establishes a connection to the PostgreSQL database.
func (p *PostgresEngine) CreatePostgresEngine(ctx context.Context) (*pgx.Conn, error) {
	err := p.assignUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("error assigning user. Err: %s", err)
	}

	// Build the connection configuration
	config, _ := pgx.ParseConfig("")
	config.Host = p.Host
	config.Port = uint16(p.Port)
	config.Database = p.Database
	config.User = p.User
	config.Password = p.Password

	// Make the connection to the PostgreSQL database
	connection, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to Database: %s, User: %s, Port: %v, Err: %s", p.Database, p.User, p.Port, err)
	}
	return connection, nil
}

func (p *PostgresEngine) assignUser(ctx context.Context) error {
	// If both user and password are missing, use IAM-based authentication
	if (p.User != "" && p.Password == "") || (p.User == "" && p.Password != "") {
		return fmt.Errorf("only one of 'user' or 'password' were specified. Either both or none should be specified")
	}
	// If neither user nor password are provided, determine the IAM email
	if p.User == "" && p.Password == "" {
		// Set user's IAMAccountEmail if available
		if p.IAMAccountEmail != "" {
			p.User = p.IAMAccountEmail
			return nil
		}
		if p.ServiceAccountRetriever != nil {
			// Otherwise, attempt to retrueve the service account email
			serviceAccountEmail, err := p.ServiceAccountRetriever.GetServiceAccountEmail(ctx)
			if err != nil {
				return fmt.Errorf("unable to retrieve service account email: %s", err)
			}
			p.User = serviceAccountEmail
		}
	}

	// If no user can be determined, return an error.
	if p.User == "" {
		return errors.New("no valid user or IAM account email provided")
	}

	return nil
}

// getServiceAccountEmail retrieves the IAM principal email with users account.
func getServiceAccountEmail(ctx context.Context) (string, error) {
	scopes := []string{"https://www.googleapis.com/auth/userinfo.email"}
	// Get credentials using email scope
	credentials, err := google.FindDefaultCredentials(ctx, scopes...) // TODO :: Can more scopes be added?
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
