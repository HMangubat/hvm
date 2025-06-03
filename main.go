package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	// Connect to PostgreSQL
	var err error
	
	dsn := "user=hvm dbname=hvmloft_shms password=PDy57XR8ZuIqwJBjHznHzQjI9VRQMEyg host=dpg-d0v74ji4d50c73e7v19g-a sslmode=disable"
	
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("DB connection error:", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("DB ping error:", err)
	}
	log.Println("✅ Connected to PostgreSQL.")
	
    // Test the DB connection
    err = db.Ping()
    if err != nil {
        log.Fatalf("Database ping failed: %v", err)
    }
    log.Println("✅ Connected to the database successfully")

	// Serve frontend static files (index.html, upload.html, scripts)
	//http.Handle("/", http.FileServer(http.Dir("./frontend")))
	fs := http.FileServer(http.Dir("./frontend"))
    http.Handle("/", fs)
	// Serve uploaded photos
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	// API routes
	http.HandleFunc("/api/upload-photo", uploadPhotoHandler)
	http.HandleFunc("/api/photos", getAllPhotosHandler)
	http.HandleFunc("/api/delete", deletePhotoHandler)
	http.HandleFunc("/api/photos/update", updatePhotoHandler)

	log.Println("🚀 Server started at http://localhost:8092")
	//log.Fatal(http.ListenAndServe(":8080", nil))
	err = http.ListenAndServe(":8092", nil)
    if err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}
func uploadPhotoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Parse error: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, "File error: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	userID := r.FormValue("user_id")
	if userID == "" {
		http.Error(w, "Missing user_id", http.StatusBadRequest)
		return
	}

	description := r.FormValue("description") // Get description from form

	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, 0755)
	}

	dstPath := filepath.Join(uploadDir, header.Filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "File save error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	io.Copy(dst, file)

	query := `INSERT INTO user_photos (user_id, filename, description) VALUES ($1, $2, $3)`
	_, err = db.Exec(query, userID, header.Filename, description)
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "✅ Uploaded and saved: %s", header.Filename)
}


func getAllPhotosHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, filename, description FROM user_photos ORDER BY id DESC")
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Photo struct {
		ID       int    `json:"id"`
		Filename string `json:"filename"`
		Description string `json:"description"`
	}

	var photos []Photo
	for rows.Next() {
		var p Photo
		if err := rows.Scan(&p.ID, &p.Filename, &p.Description); err != nil {
			http.Error(w, "DB scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		photos = append(photos, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(photos)
}

func deletePhotoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing id param", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id param", http.StatusBadRequest)
		return
	}

	// Get filename from DB
	var filename string
	err = db.QueryRow("SELECT filename FROM user_photos WHERE id=$1", id).Scan(&filename)
	if err != nil {
		http.Error(w, "Photo not found", http.StatusNotFound)
		return
	}

	// Delete DB record
	_, err = db.Exec("DELETE FROM user_photos WHERE id=$1", id)
	if err != nil {
		http.Error(w, "Failed to delete photo in DB", http.StatusInternalServerError)
		return
	}

	// Delete file from uploads folder
	err = os.Remove(filepath.Join("./uploads", filename))
	if err != nil {
		// Log but do not fail the request
		log.Println("Failed to delete file:", err)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Deleted photo id=%d filename=%s", id, filename)
}

func updatePhotoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type UpdateRequest struct {
		ID          int    `json:"id"`
		Description string `json:"description"`
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE user_photos SET description=$1 WHERE id=$2", req.Description, req.ID)
	if err != nil {
		http.Error(w, "Failed to update: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "✅ Description updated for photo ID %d", req.ID)
}

