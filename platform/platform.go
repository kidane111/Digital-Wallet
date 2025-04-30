package platform

import (
	"context"
	"net/http"
)

type HTTPClient interface {
	DoRequest(ctx context.Context, method, url string, contentTypeAccept string, modifyRequest func(*http.Request),
		body interface{}, response interface{}) (*http.Response, error)
}
