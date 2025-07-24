# Gunakan image resmi Go versi 1.22
FROM golang:1.22

# Buat direktori kerja dalam container
WORKDIR /app

# Salin file dependency
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Salin seluruh isi folder backend (termasuk main.go)
COPY . .

# Build binary Go
RUN go build -o main .

# Jalankan program
CMD ["./main"]
