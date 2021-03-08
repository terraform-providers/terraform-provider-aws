package tfresource

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

type EmptyResultError struct {
	LastRequest interface{}
}

var ErrEmptyResult = &EmptyResultError{}

func NewEmptyResultError(lastRequest interface{}) error {
	return &EmptyResultError{
		LastRequest: lastRequest,
	}
}

func (e *EmptyResultError) Error() string {
	return "empty result"
}

func (e *EmptyResultError) Is(err error) bool {
	_, ok := err.(*EmptyResultError)
	return ok
}

func (e *EmptyResultError) As(target interface{}) bool {
	t, ok := target.(**resource.NotFoundError)
	if !ok {
		return false
	}

	*t = &resource.NotFoundError{
		Message:     e.Error(),
		LastRequest: e.LastRequest,
	}

	return true
}

type TooManyResultsError struct {
	Count       int
	LastRequest interface{}
}

var ErrTooManyResults = &TooManyResultsError{}

func NewTooManyResultsError(count int, lastRequest interface{}) error {
	return &TooManyResultsError{
		Count:       count,
		LastRequest: lastRequest,
	}
}

func (e *TooManyResultsError) Error() string {
	return fmt.Sprintf("too many results: wanted 1, got %d", e.Count)
}

func (e *TooManyResultsError) Is(err error) bool {
	_, ok := err.(*TooManyResultsError)
	return ok
}

func (e *TooManyResultsError) As(target interface{}) bool {
	t, ok := target.(**resource.NotFoundError)
	if !ok {
		return false
	}

	*t = &resource.NotFoundError{
		Message:     e.Error(),
		LastRequest: e.LastRequest,
	}

	return true
}
