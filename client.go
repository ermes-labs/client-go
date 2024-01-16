package ermes_client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const DefaultTokenHeaderName = "X-ErmesSessionToken"

type ErmesToken struct {
	SessionID string `json:"sessionId"`
	Host      string `json:"host"`
}

type ErmesClientOptions struct {
	HttpClient      *http.Client
	TokenHeaderName string
	Scheme          string
	InitialOrigin   string
	InitialToken    *ErmesToken
}

type ErmesClient struct {
	httpClient         http.Client
	scheme             string
	tokenHeaderName    string
	tokenOrInitialHost interface{}
}

func NewErmesClient(options ErmesClientOptions) (*ErmesClient, error) {
	client := &ErmesClient{}

	if options.HttpClient != nil {
		client.httpClient = *options.HttpClient
	} else {
		client.httpClient = http.Client{}
	}

	if options.TokenHeaderName != "" {
		client.tokenHeaderName = options.TokenHeaderName
	} else {
		client.tokenHeaderName = DefaultTokenHeaderName
	}

	if options.InitialToken != nil {
		client.tokenOrInitialHost = options.InitialToken

		if options.Scheme != "" {
			client.scheme = options.Scheme
		} else {
			client.scheme = "https"
		}
	} else if options.InitialOrigin != "" {
		u, err := url.Parse(options.InitialOrigin)
		if err != nil {
			return nil, err
		}
		client.scheme = u.Scheme
		client.tokenOrInitialHost = u.Host
	} else {
		return nil, errors.New("either initialOrigin or initialToken must be set")
	}

	return client, nil
}

func (c *ErmesClient) Token() *ErmesToken {
	if token, ok := c.tokenOrInitialHost.(*ErmesToken); ok {
		return token
	} else {
		return nil
	}
}

func (c *ErmesClient) Host() string {
	if host, ok := c.tokenOrInitialHost.(string); ok {
		return host
	} else if token, ok := c.tokenOrInitialHost.(*ErmesToken); ok {
		return token.Host
	} else {
		return ""
	}
}

func (c *ErmesClient) StringURL(path string) string {
	return fmt.Sprintf("%s://%s%s", c.scheme, c.Host(), path)
}

func (c *ErmesClient) URL(path string) (*url.URL, error) {
	return url.Parse(c.StringURL(path))
}

func (c *ErmesClient) Get(path string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.StringURL(path), nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *ErmesClient) Post(path string, bodyType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", c.StringURL(path), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", bodyType)
	return c.do(req)
}

func (c *ErmesClient) Head(path string) (*http.Response, error) {
	req, err := http.NewRequest("HEAD", c.StringURL(path), nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *ErmesClient) Do(path string, options *http.Request) (*http.Response, error) {
	url, err := c.URL(path)
	if err != nil {
		return nil, err
	}

	options.URL = url

	return c.do(options)
}

func (c *ErmesClient) do(options *http.Request) (*http.Response, error) {
	if token := c.Token(); token != nil {
		if options.Header == nil {
			options.Header = http.Header{}
		}

		tokenBytes, err := json.Marshal(token)
		if err != nil {
			return nil, err
		}

		options.Header.Set(c.tokenHeaderName, string(tokenBytes))
	}

	response, err := c.httpClient.Do(options)

	if err != nil {
		return nil, err
	}

	if response.StatusCode == http.StatusOK {
		tokenHeader := response.Header.Get(c.tokenHeaderName)
		if tokenHeader != "" {
			token := &ErmesToken{}
			err := json.Unmarshal([]byte(tokenHeader), token)
			if err != nil {
				return nil, err
			}
			c.tokenOrInitialHost = token
		}
	}

	return response, nil
}
