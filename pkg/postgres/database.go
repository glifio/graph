package postgres

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/go-pg/pg/v10"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/viper"
)

var once sync.Once

type DB struct {
	pgx *pgxpool.Pool
}

// variabel Global
var db *DB

func GetInstanceDB() *DB {

	once.Do(func() {
		db = &DB{}

		pgx, err := pgxpool.Connect(context.Background(), viper.GetViper().GetString("lily"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to connection to database: %v\n", err)
			db.pgx = nil
			return
		}
		log.Println("db -> connected")
		db.pgx = pgx
	})

	return db
}

func (db *DB) Close() {
	if db != nil {
		db.pgx.Close()
	}
}

// old stuff

type Database struct {
	Db *pg.DB
}

// New makes a new database using the connection string and
// returns it, otherwise returns the error
func (t *Database) New(host string, port int, user string, password string, dbName string, searchPath string) error {
	//	fmt.Print(opt.TLSConfig)
	t.Db = pg.Connect(&pg.Options{
		Addr:      fmt.Sprintf("%s:%d", host, port),
		User:      user,
		Password:  password,
		Database:  dbName,
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
		return errors.New("no database connection")
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
