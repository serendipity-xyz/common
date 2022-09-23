package storage

import (
	"errors"
	"fmt"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
)

type MissingRequiredParameterError struct{}

func (e MissingRequiredParameterError) Error() string {
	return "missing or unset required parameter"
}

type NotFoundError struct{}

func (e NotFoundError) Error() string {
	return "record not found"
}

func IsNotFoundErr(err error) bool {
	if _, ok := err.(NotFoundError); ok {
		return true
	}
	return false
}

type CollisionError struct {
	CollectionName string
}

func (e CollisionError) Error() string {
	return fmt.Sprintf("collision inserting into %s", e.CollectionName)
}

func (e CollisionError) Code() int {
	return http.StatusInternalServerError
}

func isCollisionErr(err error) bool {
	var e mongo.WriteException
	if errors.As(err, &e) {
		for _, we := range e.WriteErrors {
			if we.Code == 11000 {
				return true
			}
		}
	}
	return false
}
