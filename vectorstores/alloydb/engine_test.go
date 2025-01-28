package alloydb

import (
	"context"
	"errors"
	"testing"
)

func TestGetUser(t *testing.T) {
	t.Parallel()

	testServiceAccount := "test-service-account-email@test.com"
	// Mock EmailRetriever function for testing
	var mockEmailRetrevier = func(ctx context.Context) (string, error) {
		return testServiceAccount, nil
	}

	// A failing mock function for testing
	var mockFailingEmailRetrevier = func(ctx context.Context) (string, error) {
		return "", errors.New("missing or invalid credentials")
	}

	tests := []struct {
		name             string
		engineConfig     engineConfig
		expectedErr      string
		expectedUserName string
		expectedIamAuth  bool
	}{
		{
			name:             "User and Password provided",
			engineConfig:     engineConfig{user: "testUser", password: "testPass"},
			expectedUserName: "testUser",
			expectedIamAuth:  false,
		},
		{
			name:             "Neither User nor Password, but service account email retrieved",
			engineConfig:     engineConfig{emailRetreiver: mockEmailRetrevier},
			expectedUserName: testServiceAccount,
			expectedIamAuth:  true,
		},
		{
			name:         "Error - User provided but Password missing",
			engineConfig: engineConfig{user: "testUser", password: ""},
			expectedErr:  "unable to retrieve a valid username",
		},
		{
			name:         "Error - Password provided but User missing",
			engineConfig: engineConfig{user: "", password: "testPassword"},
			expectedErr:  "unable to retrieve a valid username",
		},
		{
			name:         "Error - Failure retrieving service account email",
			engineConfig: engineConfig{emailRetreiver: mockFailingEmailRetrevier},
			expectedErr:  "unable to retrieve service account email: missing or invalid credentials",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			username, usingIAMAuth, err := getUser(context.Background(), tc.engineConfig)

			// Check if the error matches the expected error
			if err != nil && err.Error() != tc.expectedErr {
				t.Errorf("expected error %v, got %v", tc.expectedErr, err)
			}
			// If error was expected and matched, go to next test
			if tc.expectedErr != "" {
				return
			}
			// Validate if the username matches the expected username
			if username != tc.expectedUserName {
				t.Errorf("expected user %s, got %s", tc.expectedUserName, tc.engineConfig.user)
			}
			// Validate if IamAuth was expected
			if usingIAMAuth != tc.expectedIamAuth {
				t.Errorf("expected user %s, got %s", tc.expectedUserName, tc.engineConfig.user)
			}
		})
	}
}
