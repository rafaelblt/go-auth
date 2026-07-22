package api

import (
	"net/http"
)

func recovery(next http.Handler, w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			if rec == http.ErrAbortHandler {
				panic(rec)
			}

			loggerFrom(r.Context()).Error("panic recovered", "panic", rec)
			writeJSON(r.Context(), w, internalServerError())
		}
	}()
	next.ServeHTTP(w, r)
}
