package response

import (
	"encoding/json"
	"net/http"

	"workspace/pkg/apperr"
)

type Envelope struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Error   *ErrorBody `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{Success: status < 400, Data: data})
}

func Error(w http.ResponseWriter, err error) {
	ae, ok := apperr.As(err)
	if !ok {
		ae = apperr.ErrInternal
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(ae.Status)
	_ = json.NewEncoder(w).Encode(Envelope{
		Success: false,
		Error: &ErrorBody{
			Code:    ae.Code,
			Message: ae.Message,
			Details: ae.Details,
		},
	})
}

func Decode(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return apperr.ErrValidation.WithMessage("invalid JSON body")
	}
	return nil
}
