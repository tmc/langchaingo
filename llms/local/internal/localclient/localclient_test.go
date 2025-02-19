package localclient_test

import (
	"context"
	"sync"
	"testing"

	"github.com/Irooniam/langchaingo/llms/local/internal/localclient"
)

func TestBadBinPath(t *testing.T) {
	var wg sync.WaitGroup
	var err error
	var prompt string = "ABCDEF012345"
	c, err := localclient.New("/bin/ech", true)
	if err != nil {
		t.Errorf("was expecting no errors for new but got: %s", err)
	}

	r := localclient.CompletionRequest{Prompt: prompt}
	go c.CreateStreamCompletion(context.Background(), &r)
	wg.Add(1)

loop:
	for {
		select {
		case <-c.OutCh:
		case err = <-c.ErrCh:
			t.Logf("Got error %s but it was expected", err)
			break loop
		case <-c.DoneCh:
			break loop
		}
	}

	if err == nil {
		t.Error("was expecting error but error is nil")
	}

	wg.Done()

}
func TestHappyPath(t *testing.T) {
	var wg sync.WaitGroup
	var prompt string = "ABCDEF012345"
	c, err := localclient.New("/bin/echo", true)
	if err != nil {
		t.Fatalf("was not expecting errors but got: %s", err)
	}

	r := localclient.CompletionRequest{Prompt: prompt}
	go c.CreateStreamCompletion(context.Background(), &r)
	wg.Add(1)

	var msg string
loop:
	for {
		select {
		case response := <-c.OutCh:
			msg += response

		case err := <-c.ErrCh:
			t.Fatalf("wasnt expecting error from err channel for executing path but got: %s", err)
			break loop
		case <-c.DoneCh:
			break loop
		}
	}

	//response will have trailing space
	if prompt[:12] != msg[:12] {
		t.Errorf("expected respone to be '%s' but got '%s'", prompt, msg)
	}

	t.Logf("prompt %s matches response %s", prompt[:12], msg[:12])
	wg.Done()

}
