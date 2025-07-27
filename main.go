package main // Mendefinisikan package utama, yang akan menjadi entry point program.

import ( // Mengimpor package eksternal dan standar yang digunakan.
	"database/sql"  // Digunakan untuk mengakses database SQL.
	"encoding/json" // Digunakan untuk encode/decode data JSON.
	"fmt"           // Digunakan untuk format dan mencetak output.
	"log"           // Digunakan untuk logging error.
	"net/http"      // Digunakan untuk membuat server HTTP.

	_ "github.com/lib/pq" // Mengimpor driver PostgreSQL (import kosong untuk inisialisasi driver).
)

type Todo struct { // Struct untuk merepresentasikan data todo.
	ID        int    `json:"id"`        // ID todo, akan di-encode sebagai JSON dengan key "id".
	Name      string `json:"name"`      // Judul todo, key JSON "title".
	Completed bool   `json:"completed"` // Status todo, key JSON "completed".
}

var db *sql.DB // Variabel global untuk koneksi database.

func enableCORS(w http.ResponseWriter) { // Fungsi untuk mengaktifkan CORS pada response.
	w.Header().Set("Access-Control-Allow-Origin", "*")                   // Mengizinkan semua origin.
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS") // Mengizinkan method tertentu.
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")       // Mengizinkan header Content-Type.
}

func main() { // Fungsi utama (entry point).
	var err error
	db, err = sql.Open("postgres", "host=localhost port=5432 user=admin_user password=admin123 dbname=admin_db sslmode=disable")
	// Membuka koneksi ke database PostgreSQL dengan parameter koneksi.
	if err != nil {
		log.Fatal(err) // Jika gagal koneksi, hentikan program dengan log error.
	}

	http.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) { // Handler endpoint /todos.
		enableCORS(w)              // Aktifkan CORS.
		if r.Method == "OPTIONS" { // Jika request adalah preflight CORS (OPTIONS).
			return // Tidak lakukan apa-apa.
		}
		getTodos(w, r) // Panggil fungsi getTodos untuk ambil data todos.
	})

	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) { // Handler endpoint /add.
		enableCORS(w)              // Aktifkan CORS.
		if r.Method == "OPTIONS" { // Jika request adalah preflight.
			return
		}
		addTodo(w, r) // Panggil fungsi addTodo untuk menambah data todo.
	})

	fmt.Println("Server running at http://localhost:8081") // Tampilkan pesan bahwa server berjalan.
	log.Fatal(http.ListenAndServe(":8081", nil))           // Jalankan server di port 8081, hentikan jika error.
}

func getTodos(w http.ResponseWriter, _ *http.Request) { // Fungsi untuk mengambil semua todos.
	rows, err := db.Query("SELECT id, name, completed FROM todos") // Query data dari tabel todos.
	if err != nil {
		http.Error(w, err.Error(), 500) // Jika error, kembalikan status 500.
		return
	}
	defer rows.Close() // Tutup hasil query setelah selesai.

	var todos []Todo  // Slice untuk menampung data todos.
	for rows.Next() { // Iterasi tiap baris hasil query.
		var t Todo
		if err := rows.Scan(&t.ID, &t.Name, &t.Completed); err != nil { // Ambil nilai tiap kolom.
			http.Error(w, err.Error(), 500)
			return
		}
		todos = append(todos, t) // Masukkan data ke slice todos.
	}

	w.Header().Set("Content-Type", "application/json") // Set response menjadi JSON.
	json.NewEncoder(w).Encode(todos)                   // Encode slice todos menjadi JSON dan kirim ke client.
}

func addTodo(w http.ResponseWriter, r *http.Request) { // Fungsi untuk menambah todo.
	var t Todo
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil { // Decode JSON dari body request ke struct Todo.
		http.Error(w, err.Error(), 400) // Jika gagal decode, kembalikan status 400.
		return
	}

	err := db.QueryRow("INSERT INTO todos(name, completed) VALUES($1, $2) RETURNING id", t.Name, t.Completed).Scan(&t.ID)
	// Insert data todo ke database dan ambil ID yang baru dimasukkan.
	if err != nil {
		http.Error(w, err.Error(), 500) // Jika error database, kembalikan status 500.
		return
	}

	w.Header().Set("Content-Type", "application/json") // Set response menjadi JSON.
	json.NewEncoder(w).Encode(t)                       // Encode struct Todo (dengan ID baru) ke JSON dan kirim ke client.
}
