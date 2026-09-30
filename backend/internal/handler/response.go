package handler

// errorResponse is the shape of every JSON error body returned by /api handlers.
type errorResponse struct {
	Error string `json:"error"`
}

// healthzResponse is the body returned by GET /api/healthz.
type healthzResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}
