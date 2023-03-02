package cbsearchx

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type HttpSearch struct {
	HttpClient *http.Client
	Logger     *zap.Logger
	UserAgent  string
	Username   string
	Password   string
}

func (h HttpSearch) Search(ctx context.Context, indexName string, query json.RawMessage, opts *SearchOptions) (*SearchResult, error) {
	if opts == nil {
		opts = &SearchOptions{}
	}

	if opts.Endpoint == "" {
		return nil, InvalidArgumentError{"endpoint cannot be empty"}
	}

	body, err := opts.toMap(indexName)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return h.execute(ctx, query, opts.Endpoint, opts.OnBehalfOf, payload)

}

func (h HttpSearch) execute(ctx context.Context, query json.RawMessage, endpoint, onBehalfOf string, payload []byte) (*SearchResult, error) {

	reqURI := endpoint + ""
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURI)
	if err != nil {
		return nil, SearchError{
			InnerError: err,
			Query:      query,
			Endpoint:   endpoint,
			IndexName:  indexName,
		}
	}

	h.HttpClient.Do(req)
}
