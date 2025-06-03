// package main

// import (
// 	"database/sql"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"log"
// 	"net/http"
// 	"os"
// 	"path/filepath"
// 	"strconv"

// 	_ "github.com/lib/pq"
// )

// type Photo struct {
// 	ID       int    `json:"id"`
// 	UserID   string `json:"user_id"`
// 	Filename string `json:"filename"`
// }

// var db *sql.DB

// func main() {
// 	var err error
// 	dsn := "user=postgres dbname=kafka password=123 host=10.9.2.30 sslmode=disable"
// 	db, err = sql.Open("postgres", dsn)
// 	if err != nil {
// 		log.Fatal("DB connection error:", err)
// 	}
// 	if err = db.Ping(); err != nil {
// 		log.Fatal("DB ping error:", err)
// 	}
// 	log.Println("✅ Connected to PostgreSQL.")

// 	fs := http.FileServer(http.Dir("./frontend"))
// 	http.Handle("/", fs)

// 	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

// 	http.HandleFunc("/api/upload-photo", uploadPhotoHandler)
// 	http.HandleFunc("/api/photos", getAllPhotosHandler)
// 	http.HandleFunc("/api/photos/delete", deletePhotoHandler)
// 	http.HandleFunc("/api/photos/update", updatePhotoHandler)

// 	log.Println("🚀 Server started at http://localhost:8080")
// 	log.Fatal(http.ListenAndServe(":8080", nil))
// }

// func uploadPhotoHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	err := r.ParseMultipartForm(10 << 20) // 10 MB
// 	if err != nil {
// 		http.Error(w, "Parse error: "+err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	file, header, err := r.FormFile("photo")
// 	if err != nil {
// 		http.Error(w, "File error: "+err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	defer file.Close()

// 	userID := r.FormValue("user_id")
// 	if userID == "" {
// 		http.Error(w, "Missing user_id", http.StatusBadRequest)
// 		return
// 	}

// 	uploadDir := "./uploads"
// 	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
// 		os.Mkdir(uploadDir, 0755)
// 	}

// 	dstPath := filepath.Join(uploadDir, header.Filename)
// 	dst, err := os.Create(dstPath)
// 	if err != nil {
// 		http.Error(w, "File save error: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	defer dst.Close()
// 	io.Copy(dst, file)

// 	query := `INSERT INTO user_photos (user_id, filename) VALUES ($1, $2)`
// 	_, err = db.Exec(query, userID, header.Filename)
// 	if err != nil {
// 		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	fmt.Fprintf(w, "✅ Uploaded and saved: %s", header.Filename)
// }

// func getAllPhotosHandler(w http.ResponseWriter, r *http.Request) {
// 	rows, err := db.Query("SELECT id, user_id, filename FROM user_photos ORDER BY id DESC")
// 	if err != nil {
// 		http.Error(w, "DB query error: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	defer rows.Close()

// 	var photos []Photo
// 	for rows.Next() {
// 		var p Photo
// 		if err := rows.Scan(&p.ID, &p.UserID, &p.Filename); err != nil {
// 			http.Error(w, "DB scan error: "+err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 		photos = append(photos, p)
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(photos)
// }

// func deletePhotoHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodDelete {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	idStr := r.URL.Query().Get("id")
// 	if idStr == "" {
// 		http.Error(w, "Missing photo ID", http.StatusBadRequest)
// 		return
// 	}
// 	id, err := strconv.Atoi(idStr)
// 	if err != nil {
// 		http.Error(w, "Invalid photo ID", http.StatusBadRequest)
// 		return
// 	}

// 	// Get filename for deleting file
// 	var filename string
// 	err = db.QueryRow("SELECT filename FROM user_photos WHERE id=$1", id).Scan(&filename)
// 	if err != nil {
// 		http.Error(w, "Photo not found", http.StatusNotFound)
// 		return
// 	}

// 	_, err = db.Exec("DELETE FROM user_photos WHERE id=$1", id)
// 	if err != nil {
// 		http.Error(w, "DB delete error: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// Delete file from uploads folder
// 	filePath := filepath.Join("./uploads", filename)
// 	os.Remove(filePath)

// 	w.WriteHeader(http.StatusOK)
// 	fmt.Fprint(w, "✅ Photo deleted")
// }

// func updatePhotoHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPut {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	type UpdateRequest struct {
// 		ID       int    `json:"id"`
// 		Filename string `json:"filename"`
// 		UserID   string `json:"user_id"`
// 	}

// 	var req UpdateRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	// Get old filename to rename file on disk
// 	var oldFilename string
// 	err := db.QueryRow("SELECT filename FROM user_photos WHERE id=$1", req.ID).Scan(&oldFilename)
// 	if err != nil {
// 		http.Error(w, "Photo not found", http.StatusNotFound)
// 		return
// 	}

// 	// Rename file on disk if filename changed
// 	if oldFilename != req.Filename {
// 		oldPath := filepath.Join("./uploads", oldFilename)
// 		newPath := filepath.Join("./uploads", req.Filename)
// 		err = os.Rename(oldPath, newPath)
// 		if err != nil {
// 			http.Error(w, "File rename error: "+err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 	}

// 	query := `UPDATE user_photos SET filename = $1, user_id = $2 WHERE id = $3`
// 	_, err = db.Exec(query, req.Filename, req.UserID, req.ID)
// 	if err != nil {
// 		http.Error(w, "DB update error: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

//		w.WriteHeader(http.StatusOK)
//		fmt.Fprintf(w, "✅ Photo updated")
//	}
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

	log.Println("🚀 Server started at http://localhost:8092")
	//log.Fatal(http.ListenAndServe(":8080", nil))
	err = http.ListenAndServe(":8092", nil)
    if err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}

// func uploadPhotoHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	err := r.ParseMultipartForm(10 << 20) // 10 MB
// 	if err != nil {
// 		http.Error(w, "Parse error: "+err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	file, header, err := r.FormFile("photo")
// 	if err != nil {
// 		http.Error(w, "File error: "+err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	defer file.Close()

// 	// Create uploads dir if missing
// 	uploadDir := "./uploads"
// 	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
// 		os.Mkdir(uploadDir, 0755)
// 	}

// 	// Save file
// 	dstPath := filepath.Join(uploadDir, header.Filename)
// 	dst, err := os.Create(dstPath)
// 	if err != nil {
// 		http.Error(w, "File save error: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	defer dst.Close()
// 	io.Copy(dst, file)

// 	// Insert filename into DB
// 	query := `INSERT INTO user_photos (filename) VALUES ($1)`
// 	_, err = db.Exec(query, header.Filename)
// 	if err != nil {
// 		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	fmt.Fprintf(w, "✅ Uploaded and saved: %s", header.Filename)
// }
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
