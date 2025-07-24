package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

var db *sql.DB

func main() {
	var err error
	// uncommet for testing api data from html
	// db, err = sql.Open("postgres", "host=db port=5432 user=todo_user password=secret123 dbname=todo_db sslmode=disable")
	// uncomment for data non api from local vs code
	db, err = sql.Open("postgres", "host=localhost port=5432 user=todo_user password=secret123 dbname=todo_db sslmode=disable")

	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/todos", getTodos)
	http.HandleFunc("/add", addTodo)

	fmt.Println("Server running at http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func getTodos(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, title, completed FROM todos")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		todos = append(todos, t)
	}

	json.NewEncoder(w).Encode(todos)
}

func addTodo(w http.ResponseWriter, r *http.Request) {
	var t Todo
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	err := db.QueryRow("INSERT INTO todos(title, completed) VALUES($1, $2) RETURNING id", t.Title, t.Completed).Scan(&t.ID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json") // tambahkan ini
	json.NewEncoder(w).Encode(t)
}
