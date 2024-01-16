package ermes_client_test

import (
	"container/list"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	ermes_client "github.com/ermes-labs/client-go"
)

func okResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	token, _ := json.Marshal(&ermes_client.ErmesToken{
		SessionID: "sessionId",
		Host:      "host",
	})

	w.Header().Set(ermes_client.DefaultTokenHeaderName, string(token))
	w.Write([]byte("body"))
}

func notFoundResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}

func serverMock() (*httptest.Server, *list.List) {
	// Ordered list of responses.
	requests := list.New()
	// Mock the window object.
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.PushBack(r)
		switch r.URL.Path {
		case "/ok":
			okResponse(w)
		default:
			notFoundResponse(w)
		}
	})), requests
}

func serverAndClientMock(options ermes_client.ErmesClientOptions) (*httptest.Server, *list.List, *ermes_client.ErmesClient) {
	// Mock the server.
	server, requests := serverMock()

	// Initialize ErmesClient with default options.
	client, err := ermes_client.NewErmesClient(ermes_client.ErmesClientOptions{
		InitialOrigin:   server.URL,
		HttpClient:      server.Client(),
		TokenHeaderName: options.TokenHeaderName,
		Scheme:          options.Scheme,
		InitialToken:    options.InitialToken,
	})

	if err != nil {
		panic(fmt.Sprintf("Unexpected error: %v", err))
	}

	return server, requests, client
}

func TestInitializeWithDefaultOptions(t *testing.T) {
	// Mock the server.
	server, requests, client := serverAndClientMock(ermes_client.ErmesClientOptions{})
	defer server.Close()

	resource := "/resource"
	_, err := client.Get(resource)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Get and remove last element
	lastRequestElement := requests.Back()
	lastRequest := lastRequestElement.Value.(*http.Request)
	requests.Remove(lastRequestElement)

	expectedURL := fmt.Sprintf("%s%s", client.Host(), resource)
	expectedHeaders := make(http.Header)

	if !reflect.DeepEqual(lastRequest.URL, expectedURL) {
		t.Errorf("Expected fetch to be called with URL %s, got %s", expectedURL, lastRequest.URL)
	}

	if !reflect.DeepEqual(lastRequest.Header, expectedHeaders) {
		t.Errorf("Expected fetch to be called with headers %v, got %v", expectedHeaders, lastRequest.Header)
	}
}

func TestInitialOrigin(t *testing.T) {
	// Mock the server.
	InitialOrigin := "https://initial-origin.com"
	server, requests, client := serverAndClientMock(ermes_client.ErmesClientOptions{
		InitialOrigin: InitialOrigin,
	})
	defer server.Close()

	resource := "/resource"
	_, err := client.Get(resource)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Get and remove last element
	lastRequestElement := requests.Back()
	lastRequest := lastRequestElement.Value.(*http.Request)
	requests.Remove(lastRequestElement)

	expectedURL := fmt.Sprintf("%s%s", InitialOrigin, resource)
	expectedHeaders := make(http.Header)

	if !reflect.DeepEqual(lastRequest.URL, expectedURL) {
		t.Errorf("Expected fetch to be called with URL %s, got %s", expectedURL, lastRequest.URL)
	}

	if !reflect.DeepEqual(lastRequest.Header, expectedHeaders) {
		t.Errorf("Expected fetch to be called with headers %v, got %v", expectedHeaders, lastRequest.Header)
	}
}

func TestCustomTokenHeaderName(t *testing.T) {
	TokenHeaderName := "X-CustomTokenHeaderName"
	server, requests, client := serverAndClientMock(ermes_client.ErmesClientOptions{
		TokenHeaderName: TokenHeaderName,
	})
	defer server.Close()

	token := &ermes_client.ErmesToken{
		SessionID: "session-id",
		Host:      "host.com",
	}
	resource := "/resource"
	_, err := client.Get(resource)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Get and remove last element
	lastRequestElement := requests.Back()
	lastRequest := lastRequestElement.Value.(*http.Request)
	requests.Remove(lastRequestElement)

	expectedHeaders := make(http.Header)
	tokenString, err := json.Marshal(token)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedHeaders.Set(TokenHeaderName, string(tokenString))

	if !reflect.DeepEqual(lastRequest.Header, expectedHeaders) {
		t.Errorf("Expected fetch to be called with headers %v, got %v", expectedHeaders, lastRequest.Header)
	}

	if !reflect.DeepEqual(client.Token(), token) {
		t.Errorf("Expected client token to be %v, got %v", token, client.Token())
	}
}

func TestReturnedTokenForSubsequentRequests(t *testing.T) {
	// Mock the server.
	InitialOrigin := "https://initial-origin.com"
	server, requests, client := serverAndClientMock(ermes_client.ErmesClientOptions{
		InitialOrigin: InitialOrigin,
	})
	defer server.Close()

	token := &ermes_client.ErmesToken{
		SessionID: "session-id",
		Host:      "host.com",
	}
	resource := "/resource"
	_, err := client.Get(resource)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !reflect.DeepEqual(client.Token(), token) {
		t.Errorf("Expected client token to be %v, got %v", token, client.Token())
	}

	_, err = client.Get(resource)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Get and remove last element
	lastRequestElement := requests.Back()
	lastRequest := lastRequestElement.Value.(*http.Request)
	requests.Remove(lastRequestElement)

	expectedURL := fmt.Sprintf("%s%s", InitialOrigin, resource)
	expectedHeaders := make(http.Header)
	tokenString, err := json.Marshal(token)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedHeaders.Set(ermes_client.DefaultTokenHeaderName, string(tokenString))

	if !reflect.DeepEqual(lastRequest.URL, expectedURL) {
		t.Errorf("Expected fetch to be called with URL %s, got %s", expectedURL, lastRequest.URL)
	}

	if !reflect.DeepEqual(lastRequest.Header, expectedHeaders) {
		t.Errorf("Expected fetch to be called with headers %v, got %v", expectedHeaders, lastRequest.Header)
	}
}
