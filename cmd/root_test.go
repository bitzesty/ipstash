package cmd

import (
    "testing"
    "net/http"
    "net/http/httptest"
)

func TestFetchIPError(t *testing.T) {
    // Create a mock HTTP server that returns an error
    mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }))
    defer mockServer.Close()

    // Set the IP fetching URL to the mock server's URL
    ipFetchURL = mockServer.URL

    // Fetch the IP
    _, err := fetchIP() 
    if err == nil {
        t.Fatalf("Expected an error when fetching IP, but got none")
    }
}
