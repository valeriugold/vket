package database

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"github.com/jmoiron/sqlx"
	"github.com/valeriugold/vket/app/shared/vlog"
	"gopkg.in/mgo.v2"
)

var (
	// BoltDB wrapper
	BoltDB *bolt.DB
	// Mongo wrapper
	Mongo *mgo.Session
	// SQL wrapper
	// SQL *sqlx.DB
	SQL WrapperSql
	// Database info
	databases Info
)

// WrapperSql is a Decorator struct for the actual sqlx.DB
type WrapperSql struct {
	theSqlx *sqlx.DB
}

// Select logs ins and outs and calls sqlx.DB.Select
func (ws WrapperSql) Select(dest interface{}, query string, args ...interface{}) error {
	vlog.Info.Printf("__sql-query::: %s\n", sqlQueryDebugString(query, args...))
	err := ws.theSqlx.Select(dest, query, args...)
	if err == nil {
		vlog.Info.Printf("__sql-result=== %v\n", dest)
	} else {
		vlog.Info.Printf("__sql-error--- %v\n", err)
	}
	return err
}

// Get logs ins and outs and calls sqlx.DB.Get
func (ws WrapperSql) Get(dest interface{}, query string, args ...interface{}) error {
	vlog.Info.Printf("__sql-query::: %s\n", sqlQueryDebugString(query, args...))
	// vlog.Info.Printf("__sql-query::: %v --- %v\n", query, args)
	err := ws.theSqlx.Get(dest, query, args...)
	if err == nil {
		vlog.Info.Printf("__sql-result=== %v\n", dest)
	} else {
		vlog.Info.Printf("__sql-error--- %v\n", err)
	}
	return err
}

// Exec logs ins and outs and calls sqlx.DB.Exec
func (ws WrapperSql) Exec(query string, args ...interface{}) (sql.Result, error) {
	vlog.Info.Printf("__sql-query::: %s\n", sqlQueryDebugString(query, args...))
	r, err := ws.theSqlx.Exec(query, args...)
	if err != nil {
		vlog.Info.Printf("__sql-error--- %v\n", err)
	}
	return r, err
}

// SQLQueryDebugString formats an sql query inlining its arguments
// The purpose is debug only - do not send this to the database!
// Sending this to the DB is unsafe and un-performant.
func sqlQueryDebugString(query string, args ...interface{}) string {
	var buffer bytes.Buffer
	nArgs := len(args)
	// Break the string by question marks, iterate over its parts and for each
	// question mark - append an argument and format the argument according to
	// it's type, taking into consideration NULL values and quoting strings.
	for i, part := range strings.Split(query, "?") {
		buffer.WriteString(part)
		if i < nArgs {
			switch a := args[i].(type) {
			case int64:
				buffer.WriteString(fmt.Sprintf("%d", a))
			case uint32:
				buffer.WriteString(fmt.Sprintf("%d", a))
			case uint64:
				buffer.WriteString(fmt.Sprintf("%d", a))
			case bool:
				buffer.WriteString(fmt.Sprintf("%t", a))
			case sql.NullBool:
				if a.Valid {
					buffer.WriteString(fmt.Sprintf("%t", a.Bool))
				} else {
					buffer.WriteString("NULL")
				}
			case sql.NullInt64:
				if a.Valid {
					buffer.WriteString(fmt.Sprintf("%d", a.Int64))
				} else {
					buffer.WriteString("NULL")
				}
			case sql.NullString:
				if a.Valid {
					buffer.WriteString(fmt.Sprintf("%q", a.String))
				} else {
					buffer.WriteString("NULL")
				}
			case sql.NullFloat64:
				if a.Valid {
					buffer.WriteString(fmt.Sprintf("%f", a.Float64))
				} else {
					buffer.WriteString("NULL")
				}
			default:
				buffer.WriteString(fmt.Sprintf("%q", a))
			}
		}
	}
	return buffer.String()
}

// Type is the type of database from a Type* constant
type Type string

const (
	// TypeBolt is BoltDB
	TypeBolt Type = "Bolt"
	// TypeMongoDB is MongoDB
	TypeMongoDB Type = "MongoDB"
	// TypeMySQL is MySQL
	TypeMySQL Type = "MySQL"
)

// Info contains the database configurations
type Info struct {
	// Database type
	Type Type
	// MySQL info if used
	MySQL MySQLInfo
	// Bolt info if used
	Bolt BoltInfo
	// MongoDB info if used
	MongoDB MongoDBInfo
}

// MySQLInfo is the details for the database connection
type MySQLInfo struct {
	Username  string
	Password  string
	Name      string
	Hostname  string
	Port      int
	Parameter string
}

// BoltInfo is the details for the database connection
type BoltInfo struct {
	Path string
}

// MongoDBInfo is the details for the database connection
type MongoDBInfo struct {
	URL      string
	Database string
}

// DSN returns the Data Source Name
func DSN(ci MySQLInfo) string {
	// Example: root:@tcp(localhost:3306)/test
	return ci.Username +
		":" +
		ci.Password +
		"@tcp(" +
		ci.Hostname +
		":" +
		fmt.Sprintf("%d", ci.Port) +
		")/" +
		ci.Name + ci.Parameter
}

// Connect to the database
func Connect(d Info) {
	var err error

	// Store the config
	databases = d

	switch d.Type {
	case TypeMySQL:
		// Connect to MySQL
		if SQL.theSqlx, err = sqlx.Connect("mysql", DSN(d.MySQL)); err != nil {
			log.Println("SQL Driver Error", err)
		}

		// Check if is alive
		if err = SQL.theSqlx.Ping(); err != nil {
			log.Println("Database Error", err)
		}
	case TypeBolt:
		// Connect to Bolt
		if BoltDB, err = bolt.Open(d.Bolt.Path, 0600, nil); err != nil {
			log.Println("Bolt Driver Error", err)
		}
	case TypeMongoDB:
		// Connect to MongoDB
		if Mongo, err = mgo.DialWithTimeout(d.MongoDB.URL, 5*time.Second); err != nil {
			log.Println("MongoDB Driver Error", err)
			return
		}

		// Prevents these errors: read tcp 127.0.0.1:27017: i/o timeout
		Mongo.SetSocketTimeout(1 * time.Second)

		// Check if is alive
		if err = Mongo.Ping(); err != nil {
			log.Println("Database Error", err)
		}
	default:
		log.Println("No registered database in config")
	}
}

// Update makes a modification to Bolt
func Update(bucketName string, key string, dataStruct interface{}) error {
	err := BoltDB.Update(func(tx *bolt.Tx) error {
		// Create the bucket
		bucket, e := tx.CreateBucketIfNotExists([]byte(bucketName))
		if e != nil {
			return e
		}

		// Encode the record
		encodedRecord, e := json.Marshal(dataStruct)
		if e != nil {
			return e
		}

		// Store the record
		if e = bucket.Put([]byte(key), encodedRecord); e != nil {
			return e
		}
		return nil
	})
	return err
}

// View retrieves a record in Bolt
func View(bucketName string, key string, dataStruct interface{}) error {
	err := BoltDB.View(func(tx *bolt.Tx) error {
		// Get the bucket
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		// Retrieve the record
		v := b.Get([]byte(key))
		if len(v) < 1 {
			return bolt.ErrInvalid
		}

		// Decode the record
		e := json.Unmarshal(v, &dataStruct)
		if e != nil {
			return e
		}

		return nil
	})

	return err
}

// Delete removes a record from Bolt
func Delete(bucketName string, key string) error {
	err := BoltDB.Update(func(tx *bolt.Tx) error {
		// Get the bucket
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		return b.Delete([]byte(key))
	})
	return err
}

// CheckConnection returns true if MongoDB is available
func CheckConnection() bool {
	if Mongo == nil {
		Connect(databases)
	}

	if Mongo != nil {
		return true
	}

	return false
}

// ReadConfig returns the database information
func ReadConfig() Info {
	return databases
}
