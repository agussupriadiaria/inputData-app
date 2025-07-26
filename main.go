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

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func main() {
	var err error
	db, err = sql.Open("postgres", "host=localhost port=5432 user=admin_user password=admin123 dbname=admin_db sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == "OPTIONS" {
			return
		}
		getTodos(w, r)
	})

	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == "OPTIONS" {
			return
		}
		addTodo(w, r)
	})

	fmt.Println("Server running at http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func getTodos(w http.ResponseWriter, _ *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}
