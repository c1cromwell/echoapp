package api

import "net/http"

// handleJWKS serves the ES256 public signing key as a JWK Set
// (/.well-known/jwks.json) so token verifiers can validate Echo-issued JWTs
// without an out-of-band key exchange. Public, cacheable, GET-only.
func (rt *Router) handleJWKS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if rt.tokenService == nil {
		WriteError(w, http.StatusServiceUnavailable, "JWKS_UNAVAILABLE", "token service not configured", r.Header.Get("X-Request-ID"))
		return
	}
	body, err := rt.tokenService.PublicJWKS()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "JWKS_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
