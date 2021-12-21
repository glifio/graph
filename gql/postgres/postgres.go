package postgres

import (
	"database/sql"
	"fmt"

	// postgres driver
	"github.com/glifio/graph/gql/model"
	_ "github.com/lib/pq"
)

var Db *sql.DB

// Db is our database struct used for interacting with the database
// type Db struct {
// 	*sql.DB
// }

// New makes a new database using the connection string and
// returns it, otherwise returns the error
func New(connString string) (error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return err
	}

	// Check that our connection is good
	if err := db.Ping(); err != nil {
		return err
	}

	Db = db;
	return nil
}

// ConnString returns a connection string based on the parameters it's given
// This would normally also contain the password, however we're not using one
func ConnString(host string, port int, user string, password string, dbName string) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=require",
		host, port, user, password, dbName,
	)
}

// User shape
type User struct {
	ID         int
	Name       string
	Age        int
	Profession string
	Friendly   bool
}

// Message shape
type Message struct {
	cid        string
	from       string
	to         string
}

// GetUsersByName is called within our user query for graphql
// func (d *Db) GetUsersByName(name string) []User {
func GetUsersByName(name string) []User {
	// Prepare query, takes a name argument, protects from sql injection
	stmt, err := Db.Prepare("SELECT * FROM users WHERE name=$1")
	if err != nil {
		fmt.Println("GetUserByName Preperation Err: ", err)
	}

	// Make query with our stmt, passing in name argument
	rows, err := stmt.Query(name)
	if err != nil {
		fmt.Println("GetUserByName Query Err: ", err)
	}

	// Create User struct for holding each row's data
	var r User
	// Create slice of Users for our response
	users := []User{}
	// Copy the columns from row into the values pointed at by r (User)
	for rows.Next() {
		if err := rows.Scan(
			&r.ID,
			&r.Name,
			&r.Age,
			&r.Profession,
			&r.Friendly,
		); err != nil {
			fmt.Println("Error scanning rows: ", err)
		}
		users = append(users, r)
	}

	return users
}


// GetMessages 
func GetMessages() []*model.Message {
	// Prepare query, takes a name argument, protects from sql injection
	stmt, err := Db.Prepare("select m.cid, m.height, m.from, m.to, m.value, m.method, m.params from parsed_messages m limit 10")
	if err != nil {
		fmt.Println("GetMessages Preperation Err: ", err)
	}

	// Make query with our stmt, passing in name argument
	rows, err := stmt.Query()
	if err != nil {
		fmt.Println("GetMessages Query Err: ", err)
	}

	// Create User struct for holding each row's data
	//var r model.Message
	// Create slice of Users for our response
	messages := []*model.Message{}
	// Copy the columns from row into the values pointed at by r (User)
	for rows.Next() {
		msg := &model.Message{}

		if err := rows.Scan(
			&msg.Cid,
			&msg.Height,
			&msg.From,
			&msg.To,
			&msg.Value,
			&msg.Method,
			&msg.Params,
		); err != nil {
			fmt.Println("Error scanning rows: ", err)
		}
		messages = append(messages, msg)
	}

	return messages
}

// GetMessagesFrom 
func GetMessagesFrom(address string) []*model.Message {
	// Prepare query, takes a name argument, protects from sql injection
	stmt, err := Db.Prepare("select m.cid, m.height, m.from, m.to, m.value, m.method, m.params from parsed_messages m where m.from = $1 limit 100")
	if err != nil {
		fmt.Println("GetMessagesFrom Preperation Err: ", err)
	}

	// Make query with our stmt, passing in name argument
	rows, err := stmt.Query(address)
	if err != nil {
		fmt.Println("GetMessagesFrom Query Err: ", err)
	}

	// Create User struct for holding each row's data
	//var r model.Message
	// Create slice of Users for our response
	messages := []*model.Message{}
	// Copy the columns from row into the values pointed at by r (User)
	for rows.Next() {
		msg := &model.Message{}
		//&msg.From := &model.Address{} 

		if err := rows.Scan(
			&msg.Cid,
			&msg.Height,
			&msg.From,
			&msg.To,
			&msg.Value,
			&msg.Method,
			&msg.Params,
		); err != nil {
			fmt.Println("Error scanning rows: ", err)
		}
		messages = append(messages, msg)
	}

	return messages
}

// GetMessagesFrom 
func GetMessagesTo(address string) []*model.Message {
	// Prepare query, takes a name argument, protects from sql injection
	stmt, err := Db.Prepare("select m.cid, m.height, m.from, m.to, m.value, m.method, m.params from parsed_messages m where m.to = $1 limit 100")
	if err != nil {
		fmt.Println("GetMessagesTo Preperation Err: ", err)
	}

	// Make query with our stmt, passing in name argument
	rows, err := stmt.Query(address)
	if err != nil {
		fmt.Println("GetMessagesTo Query Err: ", err)
	}

	// Create User struct for holding each row's data
	//var r model.Message
	// Create slice of Users for our response
	messages := []*model.Message{}
	// Copy the columns from row into the values pointed at by r (User)
	for rows.Next() {
		msg := &model.Message{}

		if err := rows.Scan(
			&msg.Cid,
			&msg.Height,
			&msg.From,
			&msg.To,
			&msg.Value,
			&msg.Method,
			&msg.Params,
		); err != nil {
			fmt.Println("Error scanning rows: ", err)
		}
		messages = append(messages, msg)
	}

	return messages
}