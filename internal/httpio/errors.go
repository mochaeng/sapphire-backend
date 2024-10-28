package httpio

import (
	"errors"
)

var (
	ErrFormParser             = errors.New("not possible to parse form")
	ErrInvalidSearchParameter = errors.New("invalid search parameter name")

	// bad request
	ErrEmptySearchParam       = errors.New("empty search parameters were passed")
	ErrInvalidSearchParamType = errors.New("invalid type value in serach parameters")
	ErrEmptyParam             = errors.New("empty parameters were passed")

	// internal server error
	ErrMarshalData        = errors.New("error while processing data")
	ErrWrongParameterType = errors.New("you are passing wrong types to the function")
)
