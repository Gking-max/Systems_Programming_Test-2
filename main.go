package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    
    "feedback-api/internal/handlers"
    
    _ "github.com/go-sql-driver/mysql"
    "github.com/gorilla/mux"
)

func main() {
    // Database connection
    dsn := "root:password@tcp(localhost:3306)/feedback_db?parseTime=true"
    
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()
    
    // Test connection
    if err := db.Ping(); err != nil {
        log.Fatal("Failed to ping database:", err)
    }
    log.Println("Database connected successfully")
    
    // Set database in handlers
    handlers.SetDB(db)
    
    // Setup router
    router := mux.NewRouter()
    
    // API routes
    router.HandleFunc("/api/feedback", handlers.CreateFeedback).Methods("POST", "OPTIONS")
    router.HandleFunc("/api/feedback", handlers.GetAllFeedback).Methods("GET", "OPTIONS")
    router.HandleFunc("/api/feedback/{id}", handlers.GetFeedbackByID).Methods("GET", "OPTIONS")
    router.HandleFunc("/api/feedback/{id}", handlers.UpdateFeedback).Methods("PUT", "OPTIONS")
    router.HandleFunc("/api/feedback/{id}", handlers.DeleteFeedback).Methods("DELETE", "OPTIONS")
    router.HandleFunc("/api/feedback/rating/{rating}", handlers.GetFeedbackByRating).Methods("GET", "OPTIONS")
    router.HandleFunc("/api/feedback/email/{email}", handlers.GetFeedbackByEmail).Methods("GET", "OPTIONS")
    router.HandleFunc("/api/feedback/stats", handlers.GetFeedbackStats).Methods("GET", "OPTIONS")
    router.HandleFunc("/health", handlers.HealthCheck).Methods("GET")
    
    // Start server
    port := ":8080"
    log.Printf("Server starting on port %s", port)
    log.Fatal(http.ListenAndServe(port, router))
}