package vmodel

import (
	"time"

	"github.com/valeriugold/vket/app/shared/database"
)

// *****************************************************************************
// User
// *****************************************************************************

// User table contains the information for each user
type User struct {
	// ObjectID  bson.ObjectId `bson:"_id"`
	ID        uint32    `db:"id" bson:"id,omitempty"` // Don't use Id, use UserID() instead for consistency with MongoDB
	FirstName string    `db:"first_name" bson:"first_name"`
	LastName  string    `db:"last_name" bson:"last_name"`
	Email     string    `db:"email" bson:"email"`
	Password  string    `db:"password" bson:"password"`
	Role      string    `db:"role" bson:"role"`
	CreatedAt time.Time `db:"created_at" bson:"created_at"`
	UpdatedAt time.Time `db:"updated_at" bson:"updated_at"`
}

// UserGetByID gets user information from ID
func UserGetByID(id uint32) (User, error) {
	var err error

	result := User{}

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Get(&result, "SELECT * FROM user WHERE ID = ? LIMIT 1", id)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

// UserByEmail gets user information from email
func UserByEmail(email string) (User, error) {
	var err error

	result := User{}

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		// err = database.SQL.Get(&result, "SELECT id, password, role, first_name FROM user WHERE email = ? LIMIT 1", email)
		err = database.SQL.Get(&result, "SELECT * FROM user WHERE email = ? LIMIT 1", email)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

// UserCreate creates user
func UserCreate(firstName, lastName, email, password, role string) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("INSERT INTO user (first_name, last_name, email, password, role) VALUES (?,?,?,?,?)",
			firstName, lastName, email, password, role)
	default:
		err = ErrCode
	}

	return standardizeError(err)
}

func UserDelete(email string) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("DELETE FROM user WHERE email = ? LIMIT 1", email)
	default:
		err = ErrCode
	}

	return standardizeError(err)
}
