package errors

import (
	"net/http"

	"github.com/joomcode/errorx"
)

// list of error namespaces
var (
	databaseError    = errorx.NewNamespace("database error").ApplyModifiers(errorx.TypeModifierOmitStackTrace)
	invalidInput     = errorx.NewNamespace("validation error").ApplyModifiers(errorx.TypeModifierOmitStackTrace)
	resourceNotFound = errorx.NewNamespace("not found").ApplyModifiers(errorx.TypeModifierOmitStackTrace)
	serverError      = errorx.NewNamespace("server error")
	httpError        = errorx.NewNamespace("http error")
	badRequest       = errorx.NewNamespace("bad request error")
	unauthorized     = errorx.NewNamespace("unauthorized").ApplyModifiers(errorx.TypeModifierOmitStackTrace)
	AccessDenied     = errorx.RegisterTrait("You are not authorized to perform the action")
	Unauthenticated  = errorx.NewNamespace("user authentication failed")
)

var (
	ErrInvalidUserInput         = errorx.NewType(invalidInput, "invalid user input")
	ErrUnableToGet              = errorx.NewType(databaseError, "unable to get")
	ErrResourceNotFound         = errorx.NewType(resourceNotFound, "resource not found")
	ErrInternalServerError      = errorx.NewType(serverError, "internal server error")
	ErrUnableToUpdate           = errorx.NewType(databaseError, "unable to update")
	ErrUnableToDelete           = errorx.NewType(databaseError, "unable to delete")
	ErrUnableToCreate           = errorx.NewType(databaseError, "unable to create")
	ErrDBDelError               = errorx.NewType(databaseError, "could not delete record")
	ErrNoRecordFound            = errorx.NewType(resourceNotFound, "no record found")
	ErrHTTPRequestPrepareFailed = errorx.NewType(httpError, "couldn't prepare http request")
	ErrBadRequest               = errorx.NewType(badRequest, "bad request error")
	ErrAuthError                = errorx.NewType(unauthorized, "you are not authorized.")
	ErrAcessError               = errorx.NewType(unauthorized, "Unauthorized", AccessDenied)
	ErrInvalidAccessToken       = errorx.NewType(Unauthenticated, "invalid token").
					ApplyModifiers(errorx.TypeModifierOmitStackTrace)
)

type ErrorType struct {
	StatusCode int
	ErrorType  *errorx.Type
}

var Error = []ErrorType{
	{
		StatusCode: http.StatusBadRequest,
		ErrorType:  ErrInvalidUserInput,
	},
	{
		StatusCode: http.StatusInternalServerError,
		ErrorType:  ErrInternalServerError,
	},
	{
		StatusCode: http.StatusInternalServerError,
		ErrorType:  ErrUnableToGet,
	},

	{
		StatusCode: http.StatusInternalServerError,
		ErrorType:  ErrUnableToDelete,
	},

	{
		StatusCode: http.StatusNotFound,
		ErrorType:  ErrResourceNotFound,
	},

	{
		StatusCode: http.StatusNotFound,
		ErrorType:  ErrNoRecordFound,
	},
	{
		StatusCode: http.StatusInternalServerError,
		ErrorType:  ErrDBDelError,
	},
	{
		StatusCode: http.StatusInternalServerError,
		ErrorType:  ErrUnableToUpdate,
	},
	{
		StatusCode: http.StatusInternalServerError,
		ErrorType:  ErrUnableToCreate,
	},
	{
		StatusCode: http.StatusInternalServerError,
		ErrorType:  ErrHTTPRequestPrepareFailed,
	},

	{
		StatusCode: http.StatusBadRequest,
		ErrorType:  ErrBadRequest,
	},
	{
		StatusCode: http.StatusUnauthorized,
		ErrorType:  ErrAuthError,
	},
	{
		StatusCode: http.StatusUnauthorized,
		ErrorType:  ErrInvalidAccessToken,
	},
	{
		StatusCode: http.StatusAccepted,
		ErrorType:  ErrAcessError,
	},
}
