package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/url"
	"testing"
)

func TestPing(t *testing.T) {
	// create new instance of application struct
	app := &application{
		errorLog: log.New(io.Discard, "", 0),
		infoLog:  log.New(io.Discard, "", 0),
	}

	// create new test server
	// ts := httptest.NewTLSServer(app.routes())
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, body := ts.get(t, "/ping")

	if code != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, code)
	}

	if string(body) != "OK" {
		t.Errorf("want body to equal %q", "OK")
	}
}

func TestShowSnippet(t *testing.T) {
	// create new instance of application struct
	app := newTestApplication(t)

	// establish new test server
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	// table-driven tests
	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody []byte
	}{
		{"Valid ID", "/snippet/1", http.StatusOK, []byte("...")},
		{"Non-existent ID", "/snippet/2", http.StatusNotFound, nil},
		{"Negative ID", "/snippet/-1", http.StatusNotFound, nil},
		{"Decimal ID", "/snippet/0.1", http.StatusNotFound, nil},
		{"String ID", "/snippet/blah", http.StatusNotFound, nil},
		{"Empty ID", "/snippet/", http.StatusNotFound, nil},
		{"Trailing slash", "/snippet/1/", http.StatusNotFound, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, tt.urlPath)

			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}

			if !bytes.Contains(body, tt.wantBody) {
				t.Errorf("want body to contain %q", tt.wantBody)
			}
		})
	}
}

func TestSignupUser(t *testing.T) {
	// create application struct with mock dependencies
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	// make GET request
	_, _, body := ts.get(t, "/user/signup")
	csrfToken := extractCSRFToken(t, body)

	// table-driven test
	tests := []struct {
		name         string
		userName     string
		userEmail    string
		userPassword string
		csrfToken    string
		wantCode     int
		wantBody     []byte
	}{
		{"Valid", "Bob", "bob@gmail.com", "password123", csrfToken, http.StatusSeeOther, nil},
		{"Empty name", "", "bob@gmail.com", "password123", csrfToken, http.StatusOK, []byte("This field cannot be blank")},
		{"Empty email", "Bob", "", "password123", csrfToken, http.StatusOK, []byte("This field cannot be blank")},
		{"Empty password", "Bob", "bob@gmail.com", "", csrfToken, http.StatusOK, []byte("This field cannot be blank")},
		{"Invalid email (missing domain)", "Bob", "bob@gmail.", "password123", csrfToken, http.StatusOK, []byte("This field is invalid")},
		{"Invalid email (missing @)", "Bob", "bobgmail", "password123", csrfToken, http.StatusOK, []byte("This field is invalid")},
		{"Invalid email (missing local-part)", "Bob", "@gmail", "password123", csrfToken, http.StatusOK, []byte("This field is invalid")},
		{"Short password", "Bob", "bob@gmail", "pwd1", csrfToken, http.StatusOK, []byte("This field is too short (min 10 characters)")},
		{"Duplicate", "Bob", "dupe@blob.com", "password123", csrfToken, http.StatusOK, []byte("Email address already in use")},
		{"Invalid CSRF token", "", "", "", "notToken", http.StatusBadRequest, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("name", tt.userName)
			form.Add("email", tt.userEmail)
			form.Add("password", tt.userPassword)
			form.Add("csrf_token", tt.csrfToken)

			code, _, body := ts.postForm(t, "/user/signup", form)

			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}

			if !bytes.Contains(body, tt.wantBody) {
				t.Errorf("want body %s to contain %q", body, tt.wantBody)
			}
		})
	}
}
