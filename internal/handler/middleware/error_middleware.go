package middleware

import (
	"fmt"
	"net/http"

	"digital-wallet/internal/constant/errors"
	response2 "digital-wallet/internal/constant/model/response"

	"github.com/joomcode/errorx"
	"github.com/spf13/viper"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func ErrorHandler() gin.HandlerFunc {
	debugMode := viper.GetBool("debug")
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			e := c.Errors[0] // FIXME: how to handle multiple errors?
			err := e.Unwrap()

			response := CastErrorResponse(err)
			if response != nil {
				er := errorx.Cast(err)
				if debugMode {
					response.Description = fmt.Sprintf("Error: %v", er)
					response.StackTrace = fmt.Sprintf("%+v", errorx.EnsureStackTrace(err))
				}
				response2.SendErrorResponse(c, response)
				return
			}

			response2.SendErrorResponse(c, &response2.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "Unknown server error",
			})
			return
		}
	}
}
func ErrorFields(err error) []response2.FieldError {
	var errs []response2.FieldError

	if data, ok := err.(validation.Errors); ok {
		for i, v := range data {
			errs = append(errs, response2.FieldError{
				Name:        i,
				Description: v.Error(),
			},
			)
		}

		return errs
	}

	return nil
}

func CastErrorResponse(err error) *response2.ErrorResponse {
	for _, e := range errors.Error {
		if errorx.IsOfType(err, e.ErrorType) {
			er := errorx.Cast(err)
			response := response2.ErrorResponse{
				Code:       e.StatusCode,
				Message:    er.Message(),
				FieldError: ErrorFields(er.Cause()),
			}
			return &response
		}
	}
	return nil
}
