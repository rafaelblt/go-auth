package api

import (
	"encoding/json"
	"net/http"
)

func recovery(next http.Handler, w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			if rec == http.ErrAbortHandler {
				panic(rec)
			}
		
			logger := loggerFrom(r.Context())
			logger.Error("panic recovered", "panic", rec)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(internalServerErrorBody)
		}
	}()
	next.ServeHTTP(w, r)
}
