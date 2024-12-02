package alloydb

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

func (m *MockServiceAccountRetriever) serviceAccountEmailGetter(ctx context.Context) (string, error) {
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

	tests := []struct {
		name             string
		engineConfig     PostgresEngineConfig
		expectedErr      string
		expectedUserName string
		expectedIamAuth  bool
	}{
		{
			name:             "User and Password provided",
			engineConfig:     PostgresEngineConfig{user: "testUser", password: "testPass"},
			expectedUserName: "testUser",
			expectedIamAuth:  false,
		},
		{
			name:             "Neither User nor Password, but service account email retrieved",
			engineConfig:     PostgresEngineConfig{serviceAccountRetriever: succesfulServiceAccount},
			expectedUserName: testServiceAccount,
			expectedIamAuth:  true,
		},
		{
			name:         "Error - User provided but Password missing",
			engineConfig: PostgresEngineConfig{user: "testUser", password: ""},
			expectedErr:  "unable to retrieve a valid username",
		},
		{
			name:         "Error - Password provided but User missing",
			engineConfig: PostgresEngineConfig{user: "", password: "testPassword"},
			expectedErr:  "unable to retrieve a valid username",
		},
		{
			name:         "Error - Failure retrieving service account email",
			engineConfig: PostgresEngineConfig{serviceAccountRetriever: failedServiceAccount},
			expectedErr:  "unable to retrieve service account email: missing or invalid credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username, usingIAMAuth, err := tt.engineConfig.assignUser(context.Background())

			// Check if the error matches the expected error
			if err != nil && err.Error() != tt.expectedErr {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
			// If error was expected and matched, go to next test
			if tt.expectedErr != "" {
				return
			}
			// Validate if the username matches the expected username
			if username != tt.expectedUserName {
				t.Errorf("expected user %s, got %s", tt.expectedUserName, tt.engineConfig.user)
			}
			// Validate if IamAuth was expected
			if usingIAMAuth != tt.expectedIamAuth {
				t.Errorf("expected user %s, got %s", tt.expectedUserName, tt.engineConfig.user)
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
