package handlers

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "strconv"
    "strings"
    "time"
    
    "github.com/gorilla/mux"
)

// Feedback struct represents the feedback model
type Feedback struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Rating    int       `json:"rating"`
    Comments  string    `json:"comments,omitempty"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// CreateFeedbackRequest represents the request body for creating feedback
type CreateFeedbackRequest struct {
    Name     string `json:"name"`
    Email    string `json:"email"`
    Rating   int    `json:"rating"`
    Comments string `json:"comments,omitempty"`
}

// UpdateFeedbackRequest represents the request body for updating feedback
type UpdateFeedbackRequest struct {
    Name     string `json:"name"`
    Email    string `json:"email"`
    Rating   int    `json:"rating"`
    Comments string `json:"comments,omitempty"`
}

// SuccessResponse represents a successful response
type SuccessResponse struct {
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    ID      int         `json:"id,omitempty"`
    Count   int         `json:"count,omitempty"`
}

// readJSON decodes JSON from request body
func readJSON(r *http.Request, dst interface{}) error {
    decoder := json.NewDecoder(r.Body)
    defer r.Body.Close()
    return decoder.Decode(dst)
}

// writeJSON encodes data to JSON and writes to response
func writeJSON(w http.ResponseWriter, status int, data interface{}) error {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    encoder := json.NewEncoder(w)
    return encoder.Encode(data)
}

// DB global variable (will be set from main)
var DB *sql.DB

// SetDB sets the database connection for handlers
func SetDB(database *sql.DB) {
    DB = database
}

// CreateFeedback handles POST /api/feedback
func CreateFeedback(w http.ResponseWriter, r *http.Request) {
    // Handle CORS preflight
    if HandleCORS(w, r) {
        return
    }
    
    // Validate content type
    if err := ValidateContentType(r); err != nil {
        errorResp := CreateErrorResponse(ErrInvalidRequest, "Content-Type must be application/json")
        writeJSON(w, GetHTTPStatus(ErrInvalidRequest), errorResp)
        return
    }
    
    // Check rate limit
    if !CheckRateLimit(r) {
        errorResp := CreateErrorResponse(ErrInvalidRequest, "Rate limit exceeded. Please try again later.")
        writeJSON(w, http.StatusTooManyRequests, errorResp)
        return
    }
    
    // Log request
    start := time.Now()
    defer func() {
        LogRequest(r, http.StatusCreated, time.Since(start))
    }()
    
    var req CreateFeedbackRequest
    
    // Use safe JSON reading with 1MB limit
    if err := SafeReadJSON(r, &req, 1024*1024); err != nil {
        errorResp := CreateErrorResponse(ErrInvalidRequest, err.Error())
        writeJSON(w, GetHTTPStatus(ErrInvalidRequest), errorResp)
        return
    }
    
    // Sanitize inputs
    req.Name = SanitizeString(req.Name)
    req.Email = SanitizeString(req.Email)
    req.Comments = SanitizeString(req.Comments)
    
    // Validate email with stricter check
    if !IsValidEmail(req.Email) {
        errorResp := CreateErrorResponse(ErrInvalidEmail, "Please provide a valid email address")
        writeJSON(w, GetHTTPStatus(ErrInvalidEmail), errorResp)
        return
    }
    
    // Validate required fields using centralized validation
    if err := ValidateFeedbackRequest(req.Name, req.Email, req.Rating); err != nil {
        errorResp := CreateErrorResponse(err, "")
        writeJSON(w, GetHTTPStatus(err), errorResp)
        return
    }
    
    // Insert into database
    query := `INSERT INTO feedback (name, email, rating, comments) VALUES (?, ?, ?, ?)`
    result, err := DB.Exec(query, req.Name, req.Email, req.Rating, req.Comments)
    if err != nil {
        dbErr := HandleDBError(err)
        errorResp := CreateErrorResponse(dbErr, err.Error())
        writeJSON(w, GetHTTPStatus(dbErr), errorResp)
        return
    }
    
    // Get the inserted ID
    id, err := result.LastInsertId()
    if err != nil {
        dbErr := HandleDBError(err)
        errorResp := CreateErrorResponse(dbErr, err.Error())
        writeJSON(w, GetHTTPStatus(dbErr), errorResp)
        return
    }
    
    // Create response
    feedback := Feedback{
        ID:        int(id),
        Name:      req.Name,
        Email:     req.Email,
        Rating:    req.Rating,
        Comments:  req.Comments,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    
    response := SuccessResponse{
        Message: "Feedback created successfully",
        Data:    feedback,
        ID:      int(id),
    }
    
    writeJSON(w, http.StatusCreated, response)
}

// GetAllFeedback handles GET /api/feedback
func GetAllFeedback(w http.ResponseWriter, r *http.Request) {
    // Handle CORS
    if HandleCORS(w, r) {
        return
    }
    
    // Log request
    start := time.Now()
    defer func() {
        LogRequest(r, http.StatusOK, time.Since(start))
    }()
    
    // Get pagination params
    limit, offset := GetPaginationParams(r)
    
    // Get sort order from query param
    sortOrder := GetQueryParam(r, "sort", "DESC")
    if sortOrder != "ASC" && sortOrder != "DESC" {
        sortOrder = "DESC"
    }
    
    query := `SELECT id, name, email, rating, COALESCE(comments, ''), created_at, updated_at 
              FROM feedback ORDER BY created_at ` + sortOrder + ` LIMIT ? OFFSET ?`
    
    rows, err := DB.Query(query, limit, offset)
    if err != nil {
        dbErr := HandleDBError(err)
        errorResp := CreateErrorResponse(dbErr, err.Error())
        writeJSON(w, GetHTTPStatus(dbErr), errorResp)
        return
    }
    defer rows.Close()
    
    var feedbacks []Feedback
    for rows.Next() {
        var f Feedback
        err := rows.Scan(&f.ID, &f.Name, &f.Email, &f.Rating, &f.Comments, &f.CreatedAt, &f.UpdatedAt)
        if err != nil {
            dbErr := HandleDBError(err)
            errorResp := CreateErrorResponse(dbErr, err.Error())
            writeJSON(w, GetHTTPStatus(dbErr), errorResp)
            return
        }
        
        // Mask email for privacy if requested
        if GetQueryParam(r, "mask_email", "false") == "true" {
            f.Email = MaskEmail(f.Email)
        }
        
        feedbacks = append(feedbacks, f)
    }
    
    // Get total count
    var total int
    countQuery := `SELECT COUNT(*) FROM feedback`
    err = DB.QueryRow(countQuery).Scan(&total)
    if err != nil {
        dbErr := HandleDBError(err)
        errorResp := CreateErrorResponse(dbErr, err.Error())
        writeJSON(w, GetHTTPStatus(dbErr), errorResp)
        return
    }
    
    // Build paginated response
    response := BuildPaginatedResponse(feedbacks, total, limit, offset)
    response["message"] = "Feedback retrieved successfully"
    
    writeJSON(w, http.StatusOK, response)
}

// GetFeedbackByID handles GET /api/feedback/{id}
func GetFeedbackByID(w http.ResponseWriter, r *http.Request) {
    // Handle CORS
    if HandleCORS(w, r) {
        return
    }
    
    // Log request
    start := time.Now()
    defer func() {
        LogRequest(r, http.StatusOK, time.Since(start))
    }()
    
    vars := mux.Vars(r)
    idStr := vars["id"]
    
    id, err := strconv.Atoi(idStr)
    if err != nil {
        errorResp := CreateErrorResponse(ErrInvalidID, err.Error())
        writeJSON(w, GetHTTPStatus(ErrInvalidID), errorResp)
        return
    }
    
    // Validate ID
    if err := ValidateID(id); err != nil {
        errorResp := CreateErrorResponse(err, "")
        writeJSON(w, GetHTTPStatus(err), errorResp)
        return
    }
    
    query := `SELECT id, name, email, rating, COALESCE(comments, ''), created_at, updated_at 
              FROM feedback WHERE id = ?`
    
    var f Feedback
    err = DB.QueryRow(query, id).Scan(&f.ID, &f.Name, &f.Email, &f.Rating, &f.Comments, &f.CreatedAt, &f.UpdatedAt)
    
    if err != nil {
        dbErr := HandleDBError(err)
        errorResp := CreateErrorResponse(dbErr, err.Error())
        writeJSON(w, GetHTTPStatus(dbErr), errorResp)
        return
    }
    
    // Mask email if requested
    if GetQueryParam(r, "mask_email", "false") == "true" {
        f.Email = MaskEmail(f.Email)
    }
    
    writeJSON(w, http.StatusOK, f)
}

// UpdateFeedback handles PUT /api/feedback/{id}
func UpdateFeedback(w http.ResponseWriter, r *http.Request) {
    // Handle CORS
    if HandleCORS(w, r) {
        return
    }
    
    // Validate content type
    if err := ValidateContentType(r); err != nil {
        errorResp := CreateErrorResponse(ErrInvalidRequest, "Content-Type must be application/json")
        writeJSON(w, GetHTTPStatus(ErrInvalidRequest), errorResp)
        return
    }
    
    // Log request
    start := time.Now()
    defer func() {
        LogRequest(r, http.StatusOK, time.Since(start))
    }()
    
    vars := mux.Vars(r)
    idStr := vars["id"]
    
    id, err := strconv.Atoi(idStr)
    if err != nil {
        errorResp := CreateErrorResponse(ErrInvalidID, err.Error())
        writeJSON(w, GetHTTPStatus(ErrInvalidID), errorResp)
        return
    }
    
    // Validate ID
    if err := ValidateID(id); err != nil {
        errorResp := CreateErrorResponse(err, "")
        writeJSON(w, GetHTTPStatus(err), errorResp)
        return
    }
    
    var req UpdateFeedbackRequest
    if err := readJSON(r, &req); err != nil {
        errorResp := CreateErrorResponse(ErrInvalidRequest, err.Error())
        writeJSON(w, GetHTTPStatus(ErrInvalidRequest), errorResp)
        return
    }
    
    // Sanitize inputs
    req.Name = SanitizeString(req.Name)
    req.Email = SanitizeString(req.Email)
    req.Comments = SanitizeString(req.Comments)
    
    // Validate email
    if !IsValidEmail(req.Email) {
        errorResp := CreateErrorResponse(ErrInvalidEmail, "Please provide a valid email address")
        writeJSON(w, GetHTTPStatus(ErrInvalidEmail), errorResp)
        return
    }
    
    // Validate required fields
    if err := ValidateFeedbackRequest(req.Name, req.Email, req.Rating); err != nil {
        errorResp := CreateErrorResponse(err, "")
        writeJSON(w, GetHTTPStatus(err), errorResp)
        return
    }
    
    // Check if feedback exists
    var exists bool
    checkQuery := `SELECT EXISTS(SELECT 1 FROM feedback WHERE id = ?)`
    err = DB.QueryRow(checkQuery, id).Scan(&exists)
    if err != nil {
        dbErr := HandleDBError(err)
        errorResp := CreateErrorResponse(dbErr, err.Error())
        writeJSON(w, GetHTTPStatus(dbErr), errorResp)
        return
    }
    
    if !exists {
        errorResp := CreateErrorResponse(ErrFeedbackNotFound, "")
        writeJSON(w, GetHTTPStatus(ErrFeedbackNotFound), errorResp)
        return
    }
    
    // Update feedback
    query := `UPDATE feedback SET name = ?, email = ?, rating = ?, comments = ? WHERE id = ?`
    _, err = DB.Exec(query, req.Name, req.Email, req.Rating, req.Comments, id)
    if err != nil {
        dbErr := HandleDBError(err)
        errorResp := CreateErrorResponse(dbErr, err.Error())
        writeJSON(w, GetHTTPStatus(dbErr), errorResp)
        return
    }
    
    response := SuccessResponse{
        Message: "Feedback updated successfully",
        ID:      id,
    }
    
    writeJSON(w, http.StatusOK, response)
}

// DeleteFeedback handles DELETE /api/feedback/{id}
func DeleteFeedback(w http.ResponseWriter, r *http.Request) {
    // Handle CORS
    if HandleCORS(w, r) {
        return
    }
    
    // Log request
    start := time.Now()
    defer func() {
        LogRequest(r, http.StatusOK, time.Since(start))
    }()
    
    vars := mux.Vars(r)
    idStr := vars["id"]
    
    id, err := strconv.Atoi(idStr)
    if err != nil {
        errorResp := CreateErrorResponse(ErrInvalidID, err.Error())
        writeJSON(w, GetHTTPStatus(ErrInvalidID), errorResp)
        return
    }
    
    // Validate ID
    if err := ValidateID(id); err != nil {
        errorResp := CreateErrorResponse(err, "")
        writeJSON(w, GetHTTPStatus(err), errorResp)
        return
    }
    
    // Check if feedback exists
    var exists bool
    checkQuery := `SELECT EXISTS(SELECT 1 FROM feedback WHERE id = ?)`
    err = DB.QueryRow(checkQuery, id).Scan(&exists)
    if err != nil {
        dbErr := HandleDBError(err)
        errorResp := CreateErrorResponse(dbErr, err.Error())
        writeJSON(w, GetHTTPStatus(dbErr), errorResp)
        return
    }
    
    if !exists {
        errorResp := CreateErrorResponse(ErrFeedbackNotFound, "")
        writeJSON(w, GetHTTPStatus(ErrFeedbackNotFound), errorResp)
        return
    }
    
    // Delete feedback
    query := `DELETE FROM feedback WHERE id = ?`
    _, err = DB.Exec(query, id)
    if err != nil {
        dbErr := HandleDBError(err)
        errorResp := CreateErrorResponse(dbErr, err.Error())
        writeJSON(w, GetHTTPStatus(dbErr), errorResp)
        return
    }
    
    response := SuccessResponse{
        Message: "Feedback deleted successfully",
        ID:      id,
    }
    
    writeJSON(w, http.StatusOK, response)
}

// HealthCheck handles GET /health
func HealthCheck(w http.ResponseWriter, r *http.Request) {
    // Check database connection
    err := DB.Ping()
    if err != nil {
        writeJSON(w, http.StatusServiceUnavailable, JSONMap(
            "status", "unhealthy",
            "error", "Database connection failed",
        ))
        return
    }
    
    writeJSON(w, http.StatusOK, JSONMap(
        "status", "healthy",
        "timestamp", time.Now().Unix(),
    ))
}

// GetFeedbackByRating handles GET /api/feedback/rating/{rating}
func GetFeedbackByRating(w http.ResponseWriter, r *http.Request) {
    // Handle CORS
    if HandleCORS(w, r) {
        return
    }
    
    // Log request
    start := time.Now()
    defer func() {
        LogRequest(r, http.StatusOK, time.Since(start))
    }()
    
    vars := mux.Vars(r)
    ratingStr := vars["rating"]
    
    rating, err := strconv.Atoi(ratingStr)
    if err != nil {
        errorResp := CreateErrorResponse(ErrInvalidRating, err.Error())
        writeJSON(w, GetHTTPStatus(ErrInvalidRating), errorResp)
        return
    }
    
    // Validate rating
    if rating < 1 || rating > 5 {
        errorResp := CreateErrorResponse(ErrInvalidRating, "Rating must be between 1 and 5")
        writeJSON(w, GetHTTPStatus(ErrInvalidRating), errorResp)
        return
    }
    
    // Get pagination params
    limit, offset := GetPaginationParams(r)
    
    query := `SELECT id, name, email, rating, COALESCE(comments, ''), created_at, updated_at 
              FROM feedback WHERE rating = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
    
    rows, err := DB.Query(query, rating, limit, offset)
    if err != nil {
        dbErr := HandleDBError(err)
        errorResp := CreateErrorResponse(dbErr, err.Error())
        writeJSON(w, GetHTTPStatus(dbErr), errorResp)
        return
    }
    defer rows.Close()
    
    var feedbacks []Feedback
    for rows.Next() {
        var f Feedback
        err := rows.Scan(&f.ID, &f.Name, &f.Email, &f.Rating, &f.Comments, &f.CreatedAt, &f.UpdatedAt)
        if err != nil {
            dbErr := HandleDBError(err)
            errorResp := CreateErrorResponse(dbErr, err.Error())
            writeJSON(w, GetHTTPStatus(dbErr), errorResp)
            return
        }
        feedbacks = append(feedbacks, f)
    }
    
    // Get total count for this rating
    var total int
    countQuery := `SELECT COUNT(*) FROM feedback WHERE rating = ?`
    err = DB.QueryRow(countQuery, rating).Scan(&total)
    if err != nil {
        dbErr := HandleDBError(err)
        errorResp := CreateErrorResponse(dbErr, err.Error())
        writeJSON(w, GetHTTPStatus(dbErr), errorResp)
        return
    }
    
    // Build paginated response
    response := BuildPaginatedResponse(feedbacks, total, limit, offset)
    response["message"] = "Feedback retrieved successfully"
    
    writeJSON(w, http.StatusOK, response)
}

// GetFeedbackByEmail handles GET /api/feedback/email/{email}
func GetFeedbackByEmail(w http.ResponseWriter, r *http.Request) {
    // Handle CORS
    if HandleCORS(w, r) {
        return
    }
    
    // Log request
    start := time.Now()
    defer func() {
        LogRequest(r, http.StatusOK, time.Since(start))
    }()
    
    vars := mux.Vars(r)
    email := vars["email"]
    
    if IsEmptyOrWhitespace(email) {
        errorResp := CreateErrorResponse(ErrEmailRequired, "")
        writeJSON(w, GetHTTPStatus(ErrEmailRequired), errorResp)
        return
    }
    
    // Sanitize email
    email = SanitizeString(email)
    
    // Get pagination params
    limit, offset := GetPaginationParams(r)
    
    query := `SELECT id, name, email, rating, COALESCE(comments, ''), created_at, updated_at 
              FROM feedback WHERE email = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
    
    rows, err := DB.Query(query, email, limit, offset)
    if err != nil {
        dbErr := HandleDBError(err)
        errorResp := CreateErrorResponse(dbErr, err.Error())
        writeJSON(w, GetHTTPStatus(dbErr), errorResp)
        return
    }
    defer rows.Close()
    
    var feedbacks []Feedback
    for rows.Next() {
        var f Feedback
        err := rows.Scan(&f.ID, &f.Name, &f.Email, &f.Rating, &f.Comments, &f.CreatedAt, &f.UpdatedAt)
        if err != nil {
            dbErr := HandleDBError(err)
            errorResp := CreateErrorResponse(dbErr, err.Error())
            writeJSON(w, GetHTTPStatus(dbErr), errorResp)
            return
        }
        
        // Mask email for privacy
        f.Email = MaskEmail(f.Email)
        
        feedbacks = append(feedbacks, f)
    }
    
    // Get total count for this email
    var total int
    countQuery := `SELECT COUNT(*) FROM feedback WHERE email = ?`
    err = DB.QueryRow(countQuery, email).Scan(&total)
    if err != nil {
        dbErr := HandleDBError(err)
        errorResp := CreateErrorResponse(dbErr, err.Error())
        writeJSON(w, GetHTTPStatus(dbErr), errorResp)
        return
    }
    
    // Build paginated response
    response := BuildPaginatedResponse(feedbacks, total, limit, offset)
    response["message"] = "Feedback retrieved successfully"
    
    writeJSON(w, http.StatusOK, response)
}

// GetFeedbackStats handles GET /api/feedback/stats
func GetFeedbackStats(w http.ResponseWriter, r *http.Request) {
    // Handle CORS
    if HandleCORS(w, r) {
        return
    }
    
    // Log request
    start := time.Now()
    defer func() {
        LogRequest(r, http.StatusOK, time.Since(start))
    }()
    
    // Get average rating
    var avgRating float64
    avgQuery := `SELECT COALESCE(AVG(rating), 0) FROM feedback`
    err := DB.QueryRow(avgQuery).Scan(&avgRating)
    if err != nil {
        dbErr := HandleDBError(err)
        errorResp := CreateErrorResponse(dbErr, err.Error())
        writeJSON(w, GetHTTPStatus(dbErr), errorResp)
        return
    }
    
    // Get total count
    var total int
    countQuery := `SELECT COUNT(*) FROM feedback`
    err = DB.QueryRow(countQuery).Scan(&total)
    if err != nil {
        dbErr := HandleDBError(err)
        errorResp := CreateErrorResponse(dbErr, err.Error())
        writeJSON(w, GetHTTPStatus(dbErr), errorResp)
        return
    }
    
    // Get rating distribution
    ratingQuery := `SELECT rating, COUNT(*) as count FROM feedback GROUP BY rating ORDER BY rating`
    rows, err := DB.Query(ratingQuery)
    if err != nil {
        dbErr := HandleDBError(err)
        errorResp := CreateErrorResponse(dbErr, err.Error())
        writeJSON(w, GetHTTPStatus(dbErr), errorResp)
        return
    }
    defer rows.Close()
    
    ratingDistribution := make(map[int]int)
    for rows.Next() {
        var rating, count int
        err := rows.Scan(&rating, &count)
        if err != nil {
            dbErr := HandleDBError(err)
            errorResp := CreateErrorResponse(dbErr, err.Error())
            writeJSON(w, GetHTTPStatus(dbErr), errorResp)
            return
        }
        ratingDistribution[rating] = count
    }
    
    // Build response
    response := JSONMap(
        "total_feedback", total,
        "average_rating", avgRating,
        "rating_distribution", ratingDistribution,
    )
    
    writeJSON(w, http.StatusOK, response)
}