package cloudsqlutil

import (
	"context"
	"errors"
	"os"
	"testing"
)

func getEnvVariables(t *testing.T) (string, string, string, string, string, string) {
	t.Helper()

	username := os.Getenv("CLOUDSQL_USERNAME")
	if username == "" {
		t.Skip("CLOUDSQL_USERNAME environment variable not set")
	}
	password := os.Getenv("CLOUDSQL_PASSWORD")
	if password == "" {
		t.Skip("CLOUDSQL_PASSWORD environment variable not set")
	}
	database := os.Getenv("CLOUDSQL_DATABASE")
	if database == "" {
		t.Skip("CLOUDSQL_DATABASE environment variable not set")
	}
	projectID := os.Getenv("CLOUDSQL_PROJECT_ID")
	if projectID == "" {
		t.Skip("CLOUSQL_PROJECT_ID environment variable not set")
	}
	region := os.Getenv("CLOUDSQL_REGION")
	if region == "" {
		t.Skip("CLOUDSQL_REGION environment variable not set")
	}
	instance := os.Getenv("CLOUDSQL_INSTANCE")
	if instance == "" {
		t.Skip("CLOUDSQL_INSTANCE environment variable not set")
	}

	return username, password, database, projectID, region, instance
}

func TestNewPostgresEngine(t *testing.T) {
	t.Parallel()
	username, password, database, projectID, region, instance := getEnvVariables(t)
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	tcs := []struct {
		desc string
		in   []Option
		err  string
	}{
		{
			desc: "Successful Engine Creation",
			in: []Option{
				WithUser(username),
				WithPassword(password),
				WithDatabase(database),
				WithCloudSQLInstance(projectID, region, instance),
			},
			err: "",
		},
		{
			desc: "Error in engine creation with missing username and password",
			in: []Option{
				WithUser(""),
				WithPassword(""),
				WithDatabase(database),
				WithCloudSQLInstance(projectID, region, instance),
			},
			err: "missing or invalid credentials",
		},
		{
			desc: "Error in engine creation with missing instance",
			in: []Option{
				WithUser(username),
				WithPassword(password),
				WithDatabase(database),
				WithCloudSQLInstance(projectID, region, ""),
			},
			err: "missing connection: provide a connection pool or connection fields",
		},
		{
			desc: "Error in engine creation with missing projectId",
			in: []Option{
				WithUser(username),
				WithPassword(password),
				WithDatabase(database),
				WithCloudSQLInstance("", region, instance),
			},
			err: "missing connection: provide a connection pool or connection fields",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			_, err := NewPostgresEngine(ctx, tc.in...)
			if err == nil && tc.err != "" {
				t.Fatalf("unexpected error: got %q, want %q", err, tc.err)
			} else {
				errStr := err.Error()
				if errStr != tc.err {
					t.Fatalf("unexpected error: got %q, want %q", errStr, tc.err)
				}
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	t.Parallel()
	testServiceAccount := "test-service-account-email@test.com"
	// Mock EmailRetriever function for testing.
	mockEmailRetrevier := func(_ context.Context) (string, error) {
		return testServiceAccount, nil
	}

	// A failing mock function for testing.
	mockFailingEmailRetrevier := func(_ context.Context) (string, error) {
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
			engineConfig:     engineConfig{emailRetriever: mockEmailRetrevier},
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
			engineConfig: engineConfig{emailRetriever: mockFailingEmailRetrevier},
			expectedErr:  "unable to retrieve service account email: missing or invalid credentials",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			username, usingIAMAuth, err := getUser(t.Context(), tc.engineConfig)

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
