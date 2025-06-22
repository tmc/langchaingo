package localclient

import (
	"context"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name         string
		binPath      string
		globalAsArgs bool
		args         []string
	}{
		{
			name:         "basic client",
			binPath:      "/usr/bin/echo",
			globalAsArgs: false,
			args:         []string{"-n"},
		},
		{
			name:         "client with global args",
			binPath:      "/usr/bin/echo",
			globalAsArgs: true,
			args:         []string{"-n", "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.binPath, tt.globalAsArgs, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if client.BinPath != tt.binPath {
				t.Errorf("expected BinPath %s, got %s", tt.binPath, client.BinPath)
			}
			if client.GlobalAsArgs != tt.globalAsArgs {
				t.Errorf("expected GlobalAsArgs %v, got %v", tt.globalAsArgs, client.GlobalAsArgs)
			}
			if len(client.Args) != len(tt.args) {
				t.Errorf("expected %d args, got %d", len(tt.args), len(client.Args))
			}
		})
	}
}

func TestCreateCompletion(t *testing.T) {
	tests := []struct {
		name    string
		binPath string
		args    []string
		prompt  string
		want    string
	}{
		{
			name:    "echo completion",
			binPath: "echo",
			args:    []string{"-n"},
			prompt:  "Hello, World!",
			want:    "Hello, World!",
		},
		{
			name:    "echo with no args",
			binPath: "echo",
			args:    []string{},
			prompt:  "Test prompt",
			want:    "Test prompt\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.binPath, false, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error creating client: %v", err)
			}

			completion, err := client.CreateCompletion(context.Background(), &CompletionRequest{
				Prompt: tt.prompt,
			})
			if err != nil {
				t.Fatalf("unexpected error creating completion: %v", err)
			}

			if completion.Text != tt.want {
				t.Errorf("expected completion text %q, got %q", tt.want, completion.Text)
			}
		})
	}
}

func TestCreateCompletionError(t *testing.T) {
	// Test with non-existent binary
	client, err := New("/non/existent/binary", false)
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}

	_, err = client.CreateCompletion(context.Background(), &CompletionRequest{
		Prompt: "test",
	})
	if err == nil {
		t.Error("expected error for non-existent binary, got nil")
	}
}

func TestCreateCompletionWithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client, err := New("sleep", false, "1") // Command that would take 1 second
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}

	_, err = client.CreateCompletion(ctx, &CompletionRequest{
		Prompt: "test",
	})
	if err == nil {
		t.Error("expected error for cancelled context, got nil")
	}
}
