package vmodel

import (
	"database/sql"
	"errors"
	"strings"

	"gopkg.in/mgo.v2"
)

var (
	// ErrCode is a config or an internal error
	ErrCode = errors.New("Case statement in code is not correct.")
	// ErrNoResult is a not results error
	ErrNoResult = errors.New("Result not found.")
	// ErrUnavailable is a database not available error
	ErrUnavailable = errors.New("Database is unavailable.")
	// ErrUnauthorized is a permissions violation
	ErrUnauthorized = errors.New("User does not have permission to perform this operation.")
)

// standardizeErrors returns the same error regardless of the database used
func standardizeError(err error) error {
	if err == sql.ErrNoRows || err == mgo.ErrNotFound {
		return ErrNoResult
	}

	return err
}

// IsDuplicateEntry returns true if the error os "Duplicate entry"
func IsDuplicateEntry(err error) bool {
	// ERROR 1062 (23000): Duplicate entry 'bbb@aaa.aaa' for key 'email'
	if strings.Contains(err.Error(), "Duplicate entry") {
		return true
	}
	return false
}

// if driverErr, ok := err.(*mysql.MySQLError); ok { // Now the error number is accessible directly
// 	if driverErr.Number == 1045 {
// 		// Handle the permission-denied error
// 	}
// }
