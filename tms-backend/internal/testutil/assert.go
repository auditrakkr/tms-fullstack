package testutil

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

// AssertStatus fails the test if the response recorder's code doesn't match.
func AssertStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Errorf("status: got %d, want %d — body: %s", w.Code, want, w.Body.String())
	}
}

// AssertBodyContains fails if the JSON body does not contain the given key.
func AssertBodyContains(t *testing.T, w *httptest.ResponseRecorder, key string) {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if _, ok := body[key]; !ok {
		t.Errorf("response body missing key %q — body: %s", key, w.Body.String())
	}
}

// AssertNoError fails if err is not nil.
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// AssertError fails if err is nil.
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error but got nil")
	}
}
