package postgres

// Db is our database struct used for interacting with the database
import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	// postgres driver

	"github.com/go-pg/pg/v10"
	_ "github.com/lib/pq"
)

type Database struct {
 	Db 	*pg.DB
}


// New makes a new database using the connection string and
// returns it, otherwise returns the error
func (t *Database) New(host string, port int, user string, password string, dbName string, searchPath string) (error) {

//	opt, _ := pg.ParseURL("postgresql://glif-jschwartz:fXvkDRg9Y2roYeehrh@read.lilium.sh:13573/mainnet")

//	fmt.Print(opt.TLSConfig)

	t.Db = pg.Connect(&pg.Options{
		Addr: fmt.Sprintf("%s:%d", host, port),
		User: user,
		Password: password,
		Database: dbName,
		TLSConfig: &tls.Config{InsecureSkipVerify: true},
		OnConnect: func(ctx context.Context, conn *pg.Conn) error {
			_, err := conn.Exec("set search_path=?", searchPath)
			if err != nil {
				return err
			}
			return nil
		},
	})
	//db, err := sql.Open("postgres", connString)

	if t.Db == nil {
		return  errors.New("no database connection")
	}

	// Check that our connection is good
	if err := t.Db.Ping(context.Background()); err != nil {
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
