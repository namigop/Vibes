package main

import (
	"errors"
)

// package level sentinel error
var NotFoundError = errors.New("You're missing bro")

type ValidationError struct {
	Message string
}

func (v *ValidationError) Error() string {
	return v.Message
}

func Find(index int) (Shape, error) {
	if index == 0 {
		return nil, &ValidationError{Message: "Index cannot be zero. boohoo"}
	}

	// Implementation for finding a shape by index
	return nil, NotFoundError
}
