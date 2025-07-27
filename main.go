package main // mendefinisikan paket utama (entry point aplikasi)

import ( // blok import paket yang dibutuhkan
	"database/sql"  // paket standar untuk berinteraksi dengan DB via interface SQL
	"encoding/json" // untuk encoding/decoding JSON
	"fmt"           // untuk output format (Println, dll)
	"log"           // untuk logging error/fatal
	"net/http"      // untuk membuat HTTP server

	_ "github.com/lib/pq" // driver PostgreSQL; underscore untuk import side-effect (register driver)
)

type Todo struct { // struct yang merepresentasikan satu record todo
	ID        int    `json:"id"`        // kolom id, akan di-serialize sebagai "id"
	Title     string `json:"title"`     // kolom title, akan di-serialize sebagai "title"
	Completed bool   `json:"completed"` // kolom completed, akan di-serialize sebagai "completed"
}

var db *sql.DB // variabel global untuk koneksi database

func enableCORS(w http.ResponseWriter) { // helper untuk mengatur header CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")                   // izinkan semua origin
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS") // izinkan method tertentu
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")       // izinkan header Content-Type
}

func main() { // fungsi utama, start server
	var err error                                                                                                                // deklarasi variabel error
	db, err = sql.Open("postgres", "host=localhost port=5432 user=admin_user password=admin123 dbname=admin_db sslmode=disable") // buka koneksi (lazy, belum langsung tes)
	if err != nil {                                                                                                              // jika ada error saat membuka koneksi
		log.Fatal(err) // hentikan program dan log error
	}

	http.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) { // route handler untuk endpoint /todos
		enableCORS(w)              // aktifkan CORS
		if r.Method == "OPTIONS" { // tangani preflight request CORS
			return // cukup return tanpa proses lebih lanjut
		}
		getTodos(w, r) // panggil fungsi untuk mengambil semua todo
	})

	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) { // route handler untuk endpoint /add
		enableCORS(w)              // aktifkan CORS
		if r.Method == "OPTIONS" { // tangani preflight request CORS
			return // cukup return tanpa proses lebih lanjut
		}
		addTodo(w, r) // panggil fungsi untuk menambah todo
	})

	fmt.Println("Server running at http://localhost:8081") // info server berjalan
	log.Fatal(http.ListenAndServe(":8081", nil))           // jalankan HTTP server di port 8081, fatal jika error
}

func getTodos(w http.ResponseWriter, _ *http.Request) { // handler untuk mengambil list todos
	rows, err := db.Query("SELECT id, title, completed FROM todos") // eksekusi query select
	if err != nil {                                                 // jika query gagal
		http.Error(w, err.Error(), 500) // kirim error 500 ke client
		return                          // hentikan eksekusi
	}
	defer rows.Close() // pastikan rows ditutup setelah selesai dipakai

	var todos []Todo  // slice untuk menampung hasil query
	for rows.Next() { // iterasi setiap baris hasil query
		var t Todo                                                       // variabel penampung sementara
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed); err != nil { // mapping kolom ke struct
			http.Error(w, err.Error(), 500) // kirim error 500 jika scan gagal
			return                          // hentikan eksekusi
		}
		todos = append(todos, t) // masukkan todo ke slice
	}

	w.Header().Set("Content-Type", "application/json") // set response type JSON
	json.NewEncoder(w).Encode(todos)                   // encode slice todos ke JSON dan tulis ke response
}

func addTodo(w http.ResponseWriter, r *http.Request) { // handler untuk menambah todo baru
	var t Todo                                                 // struct untuk menampung body request
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil { // decode JSON dari body request
		http.Error(w, err.Error(), 400) // kirim error 400 jika JSON invalid
		return                          // hentikan eksekusi
	}

	err := db.QueryRow("INSERT INTO todos(title, completed) VALUES($1, $2) RETURNING id", t.Title, t.Completed).Scan(&t.ID) // insert data dan ambil id yang di-generate
	if err != nil {                                                                                                         // jika insert gagal
		http.Error(w, err.Error(), 500) // kirim error 500
		return                          // hentikan eksekusi
	}

	w.Header().Set("Content-Type", "application/json") // set response type JSON
	json.NewEncoder(w).Encode(t)                       // kembalikan todo yang baru dibuat (dengan id)
}
