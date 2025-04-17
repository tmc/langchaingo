package alloydbutil

import (
	"context"
	"errors"
	"os"
	"testing"
)

func getEnvVariables(t *testing.T) (string, string, string, string, string, string, string) {
	t.Helper()

	username := os.Getenv("ALLOYDB_USERNAME")
	if username == "" {
		t.Skip("ALLOYDB_USERNAME environment variable not set")
	}
	password := os.Getenv("ALLOYDB_PASSWORD")
	if password == "" {
		t.Skip("ALLOYDB_PASSWORD environment variable not set")
	}
	database := os.Getenv("ALLOYDB_DATABASE")
	if database == "" {
		t.Skip("ALLOYDB_DATABASE environment variable not set")
	}
	projectID := os.Getenv("ALLOYDB_PROJECT_ID")
	if projectID == "" {
		t.Skip("ALLOYDB_PROJECT_ID environment variable not set")
	}
	region := os.Getenv("ALLOYDB_REGION")
	if region == "" {
		t.Skip("ALLOYDB_REGION environment variable not set")
	}
	instance := os.Getenv("ALLOYDB_INSTANCE")
	if instance == "" {
		t.Skip("ALLOYDB_INSTANCE environment variable not set")
	}
	cluster := os.Getenv("ALLOYDB_CLUSTER")
	if cluster == "" {
		t.Skip("ALLOYDB_CLUSTER environment variable not set")
	}

	return username, password, database, projectID, region, instance, cluster
}

func TestNewPostgresEngine(t *testing.T) {
	t.Parallel()
	username, password, database, projectID, region, instance, cluster := getEnvVariables(t)
	ctx, cancel := context.WithCancel(context.Background())
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
				WithAlloyDBInstance(projectID, region, cluster, instance),
			},
			err: "",
		},
		{
			desc: "Error in engine creation with missing username and password",
			in: []Option{
				WithUser(""),
				WithPassword(""),
				WithDatabase(database),
				WithAlloyDBInstance(projectID, region, cluster, instance),
			},
			err: "missing or invalid credentials",
		},
		{
			desc: "Error in engine creation with missing instance",
			in: []Option{
				WithUser(username),
				WithPassword(password),
				WithDatabase(database),
				WithAlloyDBInstance("", region, cluster, instance),
			},
			err: "missing connection: provide a connection pool or connection fields",
		},
		{
			desc: "Error in engine creation with missing projectId",
			in: []Option{
				WithUser(username),
				WithPassword(password),
				WithDatabase(database),
				WithAlloyDBInstance(projectID, region, "", instance),
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
	mockEmailRetriever := func(_ context.Context) (string, error) {
		return testServiceAccount, nil
	}

	// A failing mock function for testing.
	mockFailingEmailRetriever := func(_ context.Context) (string, error) {
		return "", errors.New("missing or invalid credentials")
	}

	tests := []struct {
		name             string
		engineConfig     engineConfig
		expectedErr      string
		expectedUserName string
		expectedIAMAuth  bool
	}{
		{
			name:             "User and Password provided",
			engineConfig:     engineConfig{user: "testUser", password: "testPass"},
			expectedUserName: "testUser",
			expectedIAMAuth:  false,
		},
		{
			name:             "IAM account email provided",
			engineConfig:     engineConfig{iamAccountEmail: testServiceAccount},
			expectedUserName: testServiceAccount,
			expectedIAMAuth:  true,
		},
		{
			name:         "Getting IAM account email from the env",
			engineConfig: engineConfig{emailRetriever: mockEmailRetriever},
			expectedUserName: testServiceAccount,
			expectedIAMAuth:  true,
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
			engineConfig: engineConfig{emailRetriever: mockFailingEmailRetriever},
			expectedErr:  "unable to retrieve service account email: missing or invalid credentials",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			user, usingIAMAuth, err := getUser(context.Background(), tc.engineConfig)

			// Check if the error matches the expected error
			if err != nil && err.Error() != tc.expectedErr {
				t.Errorf("expected error %v, got %v", tc.expectedErr, err)
			}
			// If error was expected and matched, go to next test
			if tc.expectedErr != "" {
				return
			}
			// Validate if the user matches is the one expected
			if user != tc.expectedUserName {
				t.Errorf("expected user %s, got %s", tc.expectedUserName, user)
			}
			// Validate if IAMAuth was expected
			if usingIAMAuth != tc.expectedIAMAuth {
				t.Errorf("expected usingIAMAuth %t, got %t", tc.expectedIAMAuth, usingIAMAuth)
			}
		})
	}
}
