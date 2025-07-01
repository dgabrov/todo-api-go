package controller

import (
	"encoding/json"
	"fmt"
	"gitlab.com/dgb9/todo-api/internal/cst"
	"gitlab.com/dgb9/todo-api/internal/data"
	"log/slog"
	"net/http"
)

func writeResponse(w http.ResponseWriter, s any, err error) {

	// only in this case write it
	w.Header().Set(cst.ContentTypeKey, cst.ContentTypeJson)
	var er error

	if err != nil {
		errorData := data.ErrorData{Message: err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		er = json.NewEncoder(w).Encode(errorData)
	} else {
		w.WriteHeader(http.StatusOK)
		er = json.NewEncoder(w).Encode(s)
	}

	if er != nil {
		slog.Error(fmt.Sprintf("error while encoding ContentTypeJson: %v", er))
	}
}
