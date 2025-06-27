package response

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"net/http"
	"strings"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Alias  string `json:"alias,omitempty"`
}

type AliasResponse struct {
	Response
	Alias string
}

const (
	StatusOk    = "OK"
	StatusError = "Error"
)

func OK() Response {
	return Response{
		Status: StatusOk,
	}
}

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}

// ValidationError validate request
func ValidationError(errs validator.ValidationErrors) Response {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field", err.Field()))
		case "url":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not a valid URL", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not a valid", err.Field()))
		}
	}

	return Response{
		Status: StatusError,
		Error:  strings.Join(errMsgs, ", "),
	}
}

func NewJSON(w http.ResponseWriter, _ *http.Request, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(true)

	if err := enc.Encode(v); err != nil {
		w.WriteHeader(status)
		fmt.Fprintf(w, `{"error": "failed to encode response"}`)
		return
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}

func RespOk(w http.ResponseWriter, r *http.Request, alias string) {
	NewJSON(w, r, http.StatusOK, AliasResponse{
		Response: OK(),
		Alias:    alias,
	})
}
