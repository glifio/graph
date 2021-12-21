package postgres

// Db is our database struct used for interacting with the database
import (
	"database/sql"
	"fmt"

	// postgres driver

	_ "github.com/lib/pq"
)

type Database struct {
 	Db 	*sql.DB
}


// New makes a new database using the connection string and
// returns it, otherwise returns the error
func (t *Database) New(connString string) (error) {
	var err error
	t.Db, err = sql.Open("postgres", connString)
	if err != nil {
		return err
	}

	// Check that our connection is good
	if err := t.Db.Ping(); err != nil {
		return err
	}

	return nil
}

// ConnString returns a connection string based on the parameters it's given
// This would normally also contain the password, however we're not using one
func (t *Database) ConnString(host string, port int, user string, password string, dbName string, searchPath string) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=require search_path=%s",
		host, port, user, password, dbName, searchPath,
	)
}

func (t *Database) Close() error {
	return t.Db.Close()
}
