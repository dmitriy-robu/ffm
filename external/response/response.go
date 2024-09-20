package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data"`
}

func Respond(w http.ResponseWriter, rd Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(rd.Status)

	var response struct {
		Message string      `json:"message,omitempty"`
		Data    interface{} `json:"data"`
	}

	response.Message = rd.Message
	response.Data = rd.Data

	responseJson, err := json.Marshal(response)
	if err != nil {
		slog.Error("failed to marshal response", err)
	}

	_, err = w.Write(responseJson)
	if err != nil {
		return
	}

	return
}
