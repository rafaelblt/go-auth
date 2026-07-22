package api

import (
	"context"
	"encoding/json"
	"net/http"
)

func writeJSON(ctx context.Context, w http.ResponseWriter, resp response) {
	bodyBuf, err := json.Marshal(resp.Body)
	if err != nil {
		loggerFrom(ctx).Error("json marshal failed", "error", err, "body", resp.Body)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(bodyBuf)
}
