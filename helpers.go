package handlers

import (
    "database/sql"
    "net/http"
    "strconv"
    
    "feedback-api/models"
    
    "github.com/gorilla/mux"
)

var db *sql.DB

func SetDB(database *sql.DB) {
    db = database
}

func CreateFeedback(w http.ResponseWriter, r *http.Request) {
    var f models.Feedback
    if err := parseJSON(r, &f); err != nil {
        writeJSON(w, GetHTTPStatus(ErrInvalidRequest), map[string]string{"error": "Invalid JSON"})
        return
    }
    
    if err := validateFeedback(f.Name, f.Email, f.Rating); err != nil {
        writeJSON(w, GetHTTPStatus(err), map[string]string{"error": err.Error()})
        return
    }
    
    // PostgreSQL uses $1, $2 instead of ?
    err := db.QueryRow(
        "INSERT INTO feedback (name, email, rating, comments) VALUES ($1, $2, $3, $4) RETURNING id",
        f.Name, f.Email, f.Rating, f.Comments,
    ).Scan(&f.ID)
    
    if err != nil {
        writeJSON(w, 500, map[string]string{"error": err.Error()})
        return
    }
    
    writeJSON(w, 201, f)
}

func GetAllFeedback(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, name, email, rating, COALESCE(comments,'') FROM feedback ORDER BY id DESC")
    if err != nil {
        writeJSON(w, 500, map[string]string{"error": err.Error()})
        return
    }
    defer rows.Close()
    
    var feedbacks []models.Feedback
    for rows.Next() {
        var f models.Feedback
        rows.Scan(&f.ID, &f.Name, &f.Email, &f.Rating, &f.Comments)
        feedbacks = append(feedbacks, f)
    }
    
    writeJSON(w, 200, feedbacks)
}

func UpdateFeedback(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.Atoi(mux.Vars(r)["id"])
    if err != nil {
        writeJSON(w, 400, map[string]string{"error": "Invalid ID"})
        return
    }
    
    var f models.Feedback
    if err := parseJSON(r, &f); err != nil {
        writeJSON(w, 400, map[string]string{"error": "Invalid JSON"})
        return
    }
    
    if err := validateFeedback(f.Name, f.Email, f.Rating); err != nil {
        writeJSON(w, GetHTTPStatus(err), map[string]string{"error": err.Error()})
        return
    }
    
    result, err := db.Exec("UPDATE feedback SET name=$1, email=$2, rating=$3, comments=$4 WHERE id=$5",
        f.Name, f.Email, f.Rating, f.Comments, id)
    if err != nil {
        writeJSON(w, 500, map[string]string{"error": err.Error()})
        return
    }
    
    affected, _ := result.RowsAffected()
    if affected == 0 {
        writeJSON(w, 404, map[string]string{"error": "Feedback not found"})
        return
    }
    
    writeJSON(w, 200, map[string]string{"message": "Updated successfully"})
}

func DeleteFeedback(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.Atoi(mux.Vars(r)["id"])
    if err != nil {
        writeJSON(w, 400, map[string]string{"error": "Invalid ID"})
        return
    }
    
    result, err := db.Exec("DELETE FROM feedback WHERE id=$1", id)
    if err != nil {
        writeJSON(w, 500, map[string]string{"error": err.Error()})
        return
    }
    
    affected, _ := result.RowsAffected()
    if affected == 0 {
        writeJSON(w, 404, map[string]string{"error": "Feedback not found"})
        return
    }
    
    writeJSON(w, 200, map[string]string{"message": "Deleted successfully"})
}

func GetOnlyNames(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT DISTINCT name FROM feedback ORDER BY name")
    if err != nil {
        writeJSON(w, 500, map[string]string{"error": err.Error()})
        return
    }
    defer rows.Close()
    
    var names []string
    for rows.Next() {
        var name string
        rows.Scan(&name)
        names = append(names, name)
    }
    
    writeJSON(w, 200, names)
}

func GetOnlyEmails(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT DISTINCT email FROM feedback ORDER BY email")
    if err != nil {
        writeJSON(w, 500, map[string]string{"error": err.Error()})
        return
    }
    defer rows.Close()
    
    var emails []string
    for rows.Next() {
        var email string
        rows.Scan(&email)
        emails = append(emails, email)
    }
    
    writeJSON(w, 200, emails)
}

func GetOnlyComments(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT comments FROM feedback WHERE comments IS NOT NULL AND comments != ''")
    if err != nil {
        writeJSON(w, 500, map[string]string{"error": err.Error()})
        return
    }
    defer rows.Close()
    
    var comments []string
    for rows.Next() {
        var comment string
        rows.Scan(&comment)
        comments = append(comments, comment)
    }
    
    writeJSON(w, 200, comments)
}