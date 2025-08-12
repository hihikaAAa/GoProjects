package responce

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator"
)
type Response struct{
	Status string `json:"status"` // Статус нужен всегда, поэтому без omitempty
	Error string `json:"error,omitempty"` // Если все ок, значит бывают ситуации, когда ошибка не нужна => пишем omitempty
	Alias string `json:"alias,omitempty"` // omitempty - параметр в struct-tag json, если какой то параметр пустой => в итоговом json он буедт отсутствовать
}

const(
	StatusOK = "OK"
	StatusError = "Error"
)

func OK() Response{
	return Response{
		Status : StatusOK,
	}
}

func Error(msg string) Response{
	return Response{
		Status : StatusError,
		Error: msg,
	}
}

func ValidationError(errs validator.ValidationErrors) Response {
	var errMsgs []string //Возвращать будем все ошибки валидатора

	for _, err := range errs {  // Формируем ответы
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field", err.Field()))
		case "url":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not a valid URL", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}

	return Response{
		Status: StatusError,
		Error:  strings.Join(errMsgs, ", "),
	}
}