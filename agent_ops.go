package core

import "context"

func (agent *Agent) Upsert(ctx context.Context, opts *UpsertOptions) (*UpsertResult, error) {
	return agent.crud.Upsert(ctx, opts)
}

func (agent *Agent) Get(ctx context.Context, opts *GetOptions) (*GetResult, error) {
	return agent.crud.Get(ctx, opts)
}

func (agent *Agent) Delete(ctx context.Context, opts *DeleteOptions) (*DeleteResult, error) {
	return agent.crud.Delete(ctx, opts)
}

func (agent *Agent) GetAndLock(ctx context.Context, opts *GetAndLockOptions) (*GetAndLockResult, error) {
	return agent.crud.GetAndLock(ctx, opts)
}

func (agent *Agent) SendHTTPRequest(ctx context.Context, req *HTTPRequest) (*HTTPResponse, error) {
	return agent.http.SendHTTPRequest(ctx, req)
}
