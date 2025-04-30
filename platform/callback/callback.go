package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"digital-wallet/internal/constant/state"
	"digital-wallet/platform"
	"digital-wallet/platform/logger"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type client struct {
	client *http.Client
	log    logger.Logger
}

func Init(clientconfig state.HTTPTransport, log logger.Logger) platform.HTTPClient {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = clientconfig.MaxIdleConns
	t.MaxConnsPerHost = clientconfig.MaxIdleConns
	t.MaxIdleConnsPerHost = clientconfig.MaxIdleConnsPerHost
	t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:all
	return &client{
		client: &http.Client{
			Transport: t,
			Timeout:   clientconfig.Timeout,
		},
		log: log,
	}
}
func (c *client) DoRequest(
	ctx context.Context,
	method,
	url string,
	contentTypeAccept string,
	modifyRequest func(*http.Request),
	body interface{},
	response interface{},
) (*http.Response, error) {
	// Convert body to []byte
	var reqBody []byte
	if body != nil {
		var ok bool
		reqBody, ok = body.([]byte)
		if !ok {
			// Convert body to JSON
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			reqBody = jsonBody
		}
	}
	// Create a new request object
	req, err := http.NewRequest(method, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	// Set headers
	if modifyRequest != nil {
		modifyRequest(req)
	}
	if len(req.Header["Content-Type"]) < 1 {
		req.Header.Set("Content-Type", "application/json")
	}
	// Send the request and get the response
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			c.log.Warn(ctx, "Failed to close the response body", zap.Error(err))
		}
	}()

	if response == nil {
		return resp, nil
	}

	// Create a buffer to store the response content
	var responseBodyBuffer bytes.Buffer

	// Create a TeeReader to read and capture the response content
	teeReader := io.TeeReader(resp.Body, &responseBodyBuffer)

	// Read the response body
	respBody, err := io.ReadAll(teeReader)
	if err != nil {
		return nil, err
	}

	if response != nil && contentTypeAccept == "text/json" || contentTypeAccept == "" {
		err = json.Unmarshal(respBody, response)
	} else {
		err = xml.Unmarshal(respBody, response)
	}

	return resp, err
}
