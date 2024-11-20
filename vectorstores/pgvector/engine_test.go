package pgvector

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

type MockServiceAccountRetriever struct {
	MockGetServiceAccountEmail func(ctx context.Context) (string, error)
}

func (m *MockServiceAccountRetriever) GetServiceAccountEmail(ctx context.Context) (string, error) {
	return m.MockGetServiceAccountEmail(ctx)
}

func TestAssignUser(t *testing.T) {
	testServiceAccount := "test-service-account-email@test.com"
	succesfulServiceAccount := &MockServiceAccountRetriever{
		MockGetServiceAccountEmail: func(ctx context.Context) (string, error) {
			return testServiceAccount, nil
		},
	}
	failedServiceAccount := &MockServiceAccountRetriever{
		MockGetServiceAccountEmail: func(ctx context.Context) (string, error) {
			return "", errors.New("missing or invalid credentials")
		},
	}
	emptyServiceAccount := &MockServiceAccountRetriever{
		MockGetServiceAccountEmail: func(ctx context.Context) (string, error) {
			return "", nil
		},
	}

	tests := []struct {
		name         string
		engine       PostgresEngine
		expectedErr  string
		expectedUser string
	}{
		{
			name:         "User and Password provided",
			engine:       PostgresEngine{User: "testUser", Password: "testPass"},
			expectedUser: "testUser",
		},
		{
			name:         "Neither User nor Password, but IAMAccountEmail provided",
			engine:       PostgresEngine{IAMAccountEmail: "iamAccount@test.com"},
			expectedUser: "iamAccount@test.com",
		},
		{
			name:         "Neither User nor Password, but service account email retrieved",
			engine:       PostgresEngine{ServiceAccountRetriever: succesfulServiceAccount},
			expectedUser: testServiceAccount,
		},
		{
			name:        "User provided but Password missing",
			engine:      PostgresEngine{User: "testUser", Password: ""},
			expectedErr: "only one of 'user' or 'password' were specified. Either both or none should be specified",
		},
		{
			name:        "Password provided but User missing",
			engine:      PostgresEngine{User: "", Password: "testPassword"},
			expectedErr: "only one of 'user' or 'password' were specified. Either both or none should be specified",
		},
		{
			name:        "No valid user",
			engine:      PostgresEngine{ServiceAccountRetriever: emptyServiceAccount},
			expectedErr: "no valid user or IAM account email provided",
		},
		{
			name:        "Error retrieving service account email",
			engine:      PostgresEngine{ServiceAccountRetriever: failedServiceAccount},
			expectedErr: "unable to retrieve service account email: missing or invalid credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.engine.assignUser(context.Background())

			// Check if the error matches the expected error
			if err != nil && err.Error() != tt.expectedErr {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
			// If error was expected and matched, go to next test
			if tt.expectedErr != "" {
				return
			}
			// Check if the user matches the expected user
			if tt.engine.User != tt.expectedUser {
				t.Errorf("expected user %s, got %s", tt.expectedUser, tt.engine.User)
			}
		})
	}
}

// Mocks for testing getServiceAccountEmail
type MockGoogleService struct{}

func (m *MockGoogleService) FindDefaultCredentials(ctx context.Context, scopes ...string) (*google.Credentials, error) {
	return &google.Credentials{TokenSource: nil}, nil
}

type MockOAuth2Service struct{}

func (m *MockOAuth2Service) NewService(ctx context.Context, opts ...option.ClientOption) (*oauth2.Service, error) {
	return &oauth2.Service{}, nil
}

type MockUserinfoService struct {
	MockDo func() (*oauth2.Userinfo, error)
}

func (m *MockUserinfoService) Do() (*oauth2.Userinfo, error) {
	return m.MockDo()
}
