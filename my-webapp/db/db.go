package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

type User struct {
	Uuid    int    `json:"uuid"`
	Fname  string `json:"fname"`
	Lname string `json:"lname"`
}

func InitDB() {
	var err error
	dsn := "user=postgres dbname=kafka password=123 host=10.9.2.30 sslmode=disable"
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("Failed to ping DB:", err)
	}

	log.Println("✅ Connected to PostgreSQL.")
}

func GetUsers() ([]User, error) {
	rows, err := DB.Query("SELECT uuid, fname, lname FROM userlist")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		err := rows.Scan(&u.Uuid, &u.Fname, &u.Lname)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
