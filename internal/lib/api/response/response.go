package response

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"strings"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

const (
	StatusOk    = "OK"
	StatusError = "ERROR"
)

func OK() Response {
	return Response{Status: StatusOk}
}

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}

}

func ValidationError(errs validator.ValidationErrors) Response {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprint("field ", err.Field(), " is required"))
		default:
			errMsgs = append(errMsgs, fmt.Sprint("field ", err.Field(), " is not valid"))
		}
	}

	return Response{
		Status: StatusError,
		Error:  strings.Join(errMsgs, ", "),
	}
}
