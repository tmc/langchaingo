// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httprr

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"testing/iotest"
)

func handler(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/redirect") {
		http.Error(w, "redirect me!", http.StatusNotModified)
		return
	}
	if r.Method == http.MethodGet {
		if r.Header.Get("Secret") != "key" {
			http.Error(w, "missing secret", 666)
			return
		}
	}
	if r.Method == http.MethodPost {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		if !strings.Contains(string(data), "my Secret") {
			http.Error(w, "missing body secret", 667)
			return
		}
	}
}

func always555(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "should not be making HTTP requests", 555)
}

func dropPort(r *http.Request) error {
	if r.URL.Port() != "" {
		r.URL.Host = r.URL.Host[:strings.LastIndex(r.URL.Host, ":")]
		r.Host = r.Host[:strings.LastIndex(r.Host, ":")]
	}
	return nil
}

func dropSecretHeader(r *http.Request) error {
	r.Header.Del("Secret")
	return nil
}

func hideSecretBody(r *http.Request) error {
	if r.Body != nil {
		body := r.Body.(*Body)
		body.Data = []byte("redacted")
	}
	return nil
}

func doNothing(b *bytes.Buffer) error {
	return nil
}

func doRefresh(b *bytes.Buffer) error {
	s := b.String()
	b.Reset()
	_, _ = b.WriteString(s)
	return nil
}

func TestRecordReplay(t *testing.T) {
	dir := t.TempDir()
	file := dir + "/rr"

	// 4 passes:
	//	0: create
	//	1: open
	//	2: Open with -httprecord="r+"
	//	3: Open with -httprecord=""
	for pass := range 4 {
		start := open
		h := always555
		*record = ""
		switch pass {
		case 0:
			start = create
			h = handler
		case 2:
			start = Open
			*record = "r+"
			h = handler
		case 3:
			start = Open
		}
		rr, err := start(file, http.DefaultTransport)
		if err != nil {
			t.Fatal(err)
		}
		if rr.Recording() {
			t.Log("RECORDING")
		} else {
			t.Log("REPLAYING")
		}
		rr.ScrubReq(dropPort, dropSecretHeader)
		rr.ScrubReq(hideSecretBody)
		rr.ScrubResp(doNothing, doRefresh)

		mustNewRequest := func(method, url string, body io.Reader) *http.Request {
			req, err := http.NewRequest(method, url, body)
			if err != nil {
				t.Helper()
				t.Fatal(err)
			}
			return req
		}

		mustDo := func(req *http.Request, status int) {
			resp, err := rr.Client().Do(req)
			if err != nil {
				t.Helper()
				t.Fatal(err)
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode != status {
				t.Helper()
				t.Fatalf("%v: %s\n%s", req.URL, resp.Status, body)
			}
		}

		srv := httptest.NewServer(http.HandlerFunc(h))
		defer srv.Close()

		req := mustNewRequest("GET", srv.URL+"/myrequest", nil)
		req.Header.Set("Secret", "key")
		mustDo(req, 200)

		req = mustNewRequest("POST", srv.URL+"/myrequest", strings.NewReader("my Secret"))
		mustDo(req, 200)

		req = mustNewRequest("GET", srv.URL+"/redirect", nil)
		mustDo(req, 304)

		if !rr.Recording() {
			req = mustNewRequest("GET", srv.URL+"/uncached", nil)
			resp, err := rr.Client().Do(req)
			if err == nil {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				t.Fatalf("%v: %s\n%s", req.URL, resp.Status, body)
			}
		}

		if err := rr.Close(); err != nil {
			t.Fatal(err)
		}
	}

	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "Secret") {
		t.Fatalf("rr file contains Secret:\n%s", data)
	}
}

var badResponseTrace = []byte("httprr trace v1\n" +
	"105 75\n" +
	"GET http://127.0.0.1/myrequest HTTP/1.1\r\n" +
	"Host: 127.0.0.1\r\n" +
	"User-Agent: langchaingo-httprr\r\n" +
	"\r\n" +
	"HZZP/1.1 200 OK\r\n" +
	"Date: Wed, 12 Jun 2024 13:55:02 GMT\r\n" +
	"Content-Length: 0\r\n" +
	"\r\n")

func TestErrors(t *testing.T) {
	dir := t.TempDir()
	var resp *http.Response
	var err error

	makeTmpFile := func() string {
		f, err := os.CreateTemp(dir, "TestErrors")
		if err != nil {
			t.Fatalf("failed to create tmp file for test: %v", err)
		}
		name := f.Name()
		f.Close()
		return name
	}

	// -httprecord regexp parsing
	*record = "+"
	if _, err := Open(makeTmpFile(), nil); err == nil || !strings.Contains(err.Error(), "invalid -httprecord flag") {
		t.Errorf("did not diagnose bad -httprecord: err = %v", err)
	}
	*record = ""

	// invalid httprr trace
	if _, err := Open(makeTmpFile(), nil); err == nil || !strings.Contains(err.Error(), "not an httprr trace") {
		t.Errorf("did not diagnose invalid httprr trace: err = %v", err)
	}

	// corrupt httprr trace
	corruptTraceFile := makeTmpFile()
	os.WriteFile(corruptTraceFile, []byte("httprr trace v1\ngarbage\n"), 0o666)
	if _, err := Open(corruptTraceFile, nil); err == nil || !strings.Contains(err.Error(), "corrupt httprr trace") {
		t.Errorf("did not diagnose invalid httprr trace: err = %v", err)
	}

	// os.Create error creating trace
	if _, err := create("invalid\x00file", nil); err == nil {
		t.Errorf("did not report failure from os.Create: err = %v", err)
	}

	// os.ReadAll error reading trace
	if _, err := open("nonexistent", nil); err == nil {
		t.Errorf("did not report failure from os.ReadFile: err = %v", err)
	}

	// error reading body
	rr, err := create(makeTmpFile(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp, err := rr.Client().Post("http://127.0.0.1/nonexist", "x/error", iotest.ErrReader(errors.New("MY ERROR"))); err == nil || !strings.Contains(err.Error(), "MY ERROR") {
		if resp != nil {
			resp.Body.Close()
		}
		t.Errorf("did not report failure from io.ReadAll(body): err = %v", err)
	}

	// error during request scrub
	rr.ScrubReq(func(*http.Request) error { return errors.New("SCRUB ERROR") })
	resp, err = rr.Client().Get("http://127.0.0.1/nonexist")
	if resp != nil {
		resp.Body.Close()
	}
	if err == nil || !strings.Contains(err.Error(), "SCRUB ERROR") {
		t.Errorf("did not report failure from scrub: err = %v", err)
	}
	rr.Close()

	// error during response scrub
	rr.ScrubResp(func(*bytes.Buffer) error { return errors.New("SCRUB ERROR") })
	resp2, err := rr.Client().Get("http://127.0.0.1/nonexist")
	if resp2 != nil {
		resp2.Body.Close()
	}
	if err == nil || !strings.Contains(err.Error(), "SCRUB ERROR") {
		t.Errorf("did not report failure from scrub: err = %v", err)
	}
	rr.Close()

	// error during rkey.WriteProxy
	rr, err = create(makeTmpFile(), http.DefaultTransport)
	if err != nil {
		t.Fatal(err)
	}
	rr.ScrubReq(func(req *http.Request) error {
		req.Host = ""
		return nil
	})
	resp3, err := rr.Client().Get("http://127.0.0.1/nonexist")
	if resp3 != nil {
		resp3.Body.Close()
	}
	if err == nil {
		t.Errorf("expected error from rkey.WriteProxy, got nil")
	} else if !strings.Contains(err.Error(), "no Host or URL set") && !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("did not report expected failure from rkey.WriteProxy: err = %v", err)
	}
	rr.Close()

	// error during resp.Write
	rr, err = create(makeTmpFile(), badRespTransport{})
	if err != nil {
		t.Fatal(err)
	}
	resp, err = rr.Client().Get("http://127.0.0.1/nonexist")
	if resp != nil {
		resp.Body.Close()
	}
	if err == nil || !strings.Contains(err.Error(), "TRANSPORT ERROR") {
		t.Errorf("did not report failure from resp.Write: err = %v", err)
	}
	rr.Close()

	// error during Write logging request
	srv := httptest.NewServer(http.HandlerFunc(always555))
	defer srv.Close()
	rr, err = create(makeTmpFile(), http.DefaultTransport)
	if err != nil {
		t.Fatal(err)
	}
	rr.ScrubReq(dropPort)
	rr.record.Close() // cause write error
	resp, err = rr.Client().Get(srv.URL + "/redirect")
	if resp != nil {
		resp.Body.Close()
	}
	if err == nil || !strings.Contains(err.Error(), "file already closed") {
		t.Errorf("did not report failure from record write: err = %v", err)
	}
	rr.writeErr = errors.New("BROKEN ERROR")
	resp, err = rr.Client().Get(srv.URL + "/redirect")
	if resp != nil {
		resp.Body.Close()
	}
	if err == nil || !strings.Contains(err.Error(), "BROKEN ERROR") {
		t.Errorf("did not report previous write failure: err = %v", err)
	}
	if err := rr.Close(); err == nil || !strings.Contains(err.Error(), "BROKEN ERROR") {
		t.Errorf("did not report write failure during close: err = %v", err)
	}

	// error during RoundTrip
	rr, err = create(makeTmpFile(), errTransport{errors.New("TRANSPORT ERROR")})
	if err != nil {
		t.Fatal(err)
	}
	if resp, err := rr.Client().Get(srv.URL); err == nil || !strings.Contains(err.Error(), "TRANSPORT ERROR") {
		if resp != nil {
			resp.Body.Close()
		}
		t.Errorf("did not report failure from transport: err = %v", err)
	}
	rr.Close()

	// error during http.ReadResponse: trace is structurally okay but has malformed response inside
	tmpFile := makeTmpFile()
	if err := os.WriteFile(tmpFile, badResponseTrace, 0o666); err != nil {
		t.Fatal(err)
	}
	rr, err = Open(tmpFile, nil)
	if err != nil {
		if strings.Contains(err.Error(), "corrupt httprr trace") {
			// This is actually what we want - the corrupt trace is detected during open
			return
		}
		t.Fatal(err)
	}
	if resp, err := rr.Client().Get("http://127.0.0.1/myrequest"); err == nil || !strings.Contains(err.Error(), "corrupt httprr trace:") {
		if resp != nil {
			resp.Body.Close()
		}
		t.Errorf("did not diagnose invalid httprr trace: err = %v", err)
	}
	rr.Close()
}

type errTransport struct{ err error }

func (e errTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, e.err
}

type badRespTransport struct{}

func (badRespTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp := new(http.Response)
	resp.Body = io.NopCloser(iotest.ErrReader(errors.New("TRANSPORT ERROR")))
	return resp, nil
}
