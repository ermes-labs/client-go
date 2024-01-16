package ermes_client

import (
	"context"
	"io"
	"net/http"
)

const placeholderURL = "https://google.com"

func NewErmesRequest(method string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, placeholderURL, body)
}

func NewErmesRequestWithContext(context context.Context, method string, body io.Reader) (*http.Request, error) {
	return http.NewRequestWithContext(context, method, placeholderURL, body)
}
