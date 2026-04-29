package main

import (
    "database/sql"
    "log"
    "net/http"
    
    "feedback-api/handlers"
    
    _ "github.com/lib/pq" // PostgreSQL driver
    "github.com/gorilla/mux"
)

func main() {
    // PostgreSQL connection string
    connStr := "postgres://postgres:postgres@db:5432/student_event_tracker?sslmode=disable"
    
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    if err = db.Ping(); err != nil {
        log.Fatal(err)
    }
    
    handlers.SetDB(db)
    
    r := mux.NewRouter()
    r.HandleFunc("/feedback", handlers.CreateFeedback).Methods("POST")
    r.HandleFunc("/feedback", handlers.GetAllFeedback).Methods("GET")
    r.HandleFunc("/feedback/{id}", handlers.UpdateFeedback).Methods("PUT")
    r.HandleFunc("/feedback/{id}", handlers.DeleteFeedback).Methods("DELETE")
    r.HandleFunc("/feedback/names", handlers.GetOnlyNames).Methods("GET")
    r.HandleFunc("/feedback/emails", handlers.GetOnlyEmails).Methods("GET")
    r.HandleFunc("/feedback/comments", handlers.GetOnlyComments).Methods("GET")
    
    log.Println("Server on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}