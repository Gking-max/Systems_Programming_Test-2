package models

import (
    "database/sql"
    "time"
)

// Feedback represents the feedback data model
type Feedback struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Rating    int       `json:"rating"`
    Comments  string    `json:"comments,omitempty"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// CreateFeedbackInput represents the input for creating feedback
type CreateFeedbackInput struct {
    Name     string `json:"name"`
    Email    string `json:"email"`
    Rating   int    `json:"rating"`
    Comments string `json:"comments,omitempty"`
}

// UpdateFeedbackInput represents the input for updating feedback
type UpdateFeedbackInput struct {
    Name     string `json:"name"`
    Email    string `json:"email"`
    Rating   int    `json:"rating"`
    Comments string `json:"comments,omitempty"`
}

// FeedbackResponse represents the response structure for feedback
type FeedbackResponse struct {
    Message  string    `json:"message,omitempty"`
    Feedback *Feedback `json:"feedback,omitempty"`
    ID       int       `json:"id,omitempty"`
}

// ListResponse represents the response structure for listing feedback
type ListResponse struct {
    Feedback []Feedback `json:"feedback"`
    Count    int        `json:"count"`
}

// StatsResponse represents the statistics response
type StatsResponse struct {
    TotalFeedback     int            `json:"total_feedback"`
    AverageRating     float64        `json:"average_rating"`
    RatingDistribution map[int]int   `json:"rating_distribution"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
    Error string `json:"error"`
}

// FeedbackModel handles database operations for feedback
type FeedbackModel struct {
    DB *sql.DB
}

// NewFeedbackModel creates a new FeedbackModel
func NewFeedbackModel(db *sql.DB) *FeedbackModel {
    return &FeedbackModel{DB: db}
}

// Create inserts a new feedback record into the database
func (m *FeedbackModel) Create(input *CreateFeedbackInput) (*Feedback, error) {
    query := `INSERT INTO feedback (name, email, rating, comments) VALUES (?, ?, ?, ?)`
    
    result, err := m.DB.Exec(query, input.Name, input.Email, input.Rating, input.Comments)
    if err != nil {
        return nil, err
    }
    
    id, err := result.LastInsertId()
    if err != nil {
        return nil, err
    }
    
    // Return the created feedback
    return m.GetByID(int(id))
}

// GetByID retrieves a feedback record by ID
func (m *FeedbackModel) GetByID(id int) (*Feedback, error) {
    query := `SELECT id, name, email, rating, COALESCE(comments, ''), created_at, updated_at 
              FROM feedback WHERE id = ?`
    
    var feedback Feedback
    var comments sql.NullString
    
    err := m.DB.QueryRow(query, id).Scan(
        &feedback.ID,
        &feedback.Name,
        &feedback.Email,
        &feedback.Rating,
        &comments,
        &feedback.CreatedAt,
        &feedback.UpdatedAt,
    )
    
    if err != nil {
        return nil, err
    }
    
    if comments.Valid {
        feedback.Comments = comments.String
    }
    
    return &feedback, nil
}

// GetAll retrieves all feedback records with pagination
func (m *FeedbackModel) GetAll(limit, offset int, sortOrder string) ([]Feedback, int, error) {
    // Validate sort order
    if sortOrder != "ASC" && sortOrder != "DESC" {
        sortOrder = "DESC"
    }
    
    // Get total count
    var total int
    countQuery := `SELECT COUNT(*) FROM feedback`
    err := m.DB.QueryRow(countQuery).Scan(&total)
    if err != nil {
        return nil, 0, err
    }
    
    // Get paginated results
    query := `SELECT id, name, email, rating, COALESCE(comments, ''), created_at, updated_at 
              FROM feedback ORDER BY created_at ` + sortOrder + ` LIMIT ? OFFSET ?`
    
    rows, err := m.DB.Query(query, limit, offset)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()
    
    var feedbacks []Feedback
    for rows.Next() {
        var f Feedback
        var comments sql.NullString
        
        err := rows.Scan(
            &f.ID,
            &f.Name,
            &f.Email,
            &f.Rating,
            &comments,
            &f.CreatedAt,
            &f.UpdatedAt,
        )
        if err != nil {
            return nil, 0, err
        }
        
        if comments.Valid {
            f.Comments = comments.String
        }
        
        feedbacks = append(feedbacks, f)
    }
    
    return feedbacks, total, nil
}

// Update updates an existing feedback record
func (m *FeedbackModel) Update(id int, input *UpdateFeedbackInput) error {
    query := `UPDATE feedback SET name = ?, email = ?, rating = ?, comments = ? WHERE id = ?`
    
    _, err := m.DB.Exec(query, input.Name, input.Email, input.Rating, input.Comments, id)
    return err
}

// Delete removes a feedback record by ID
func (m *FeedbackModel) Delete(id int) error {
    query := `DELETE FROM feedback WHERE id = ?`
    
    result, err := m.DB.Exec(query, id)
    if err != nil {
        return err
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    
    if rowsAffected == 0 {
        return sql.ErrNoRows
    }
    
    return nil
}

// Exists checks if a feedback record exists by ID
func (m *FeedbackModel) Exists(id int) (bool, error) {
    var exists bool
    query := `SELECT EXISTS(SELECT 1 FROM feedback WHERE id = ?)`
    err := m.DB.QueryRow(query, id).Scan(&exists)
    return exists, err
}

// GetByRating retrieves all feedback records with a specific rating
func (m *FeedbackModel) GetByRating(rating, limit, offset int) ([]Feedback, int, error) {
    // Get total count for this rating
    var total int
    countQuery := `SELECT COUNT(*) FROM feedback WHERE rating = ?`
    err := m.DB.QueryRow(countQuery, rating).Scan(&total)
    if err != nil {
        return nil, 0, err
    }
    
    // Get paginated results
    query := `SELECT id, name, email, rating, COALESCE(comments, ''), created_at, updated_at 
              FROM feedback WHERE rating = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
    
    rows, err := m.DB.Query(query, rating, limit, offset)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()
    
    var feedbacks []Feedback
    for rows.Next() {
        var f Feedback
        var comments sql.NullString
        
        err := rows.Scan(
            &f.ID,
            &f.Name,
            &f.Email,
            &f.Rating,
            &comments,
            &f.CreatedAt,
            &f.UpdatedAt,
        )
        if err != nil {
            return nil, 0, err
        }
        
        if comments.Valid {
            f.Comments = comments.String
        }
        
        feedbacks = append(feedbacks, f)
    }
    
    return feedbacks, total, nil
}

// GetByEmail retrieves all feedback records by email
func (m *FeedbackModel) GetByEmail(email string, limit, offset int) ([]Feedback, int, error) {
    // Get total count for this email
    var total int
    countQuery := `SELECT COUNT(*) FROM feedback WHERE email = ?`
    err := m.DB.QueryRow(countQuery, email).Scan(&total)
    if err != nil {
        return nil, 0, err
    }
    
    // Get paginated results
    query := `SELECT id, name, email, rating, COALESCE(comments, ''), created_at, updated_at 
              FROM feedback WHERE email = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
    
    rows, err := m.DB.Query(query, email, limit, offset)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()
    
    var feedbacks []Feedback
    for rows.Next() {
        var f Feedback
        var comments sql.NullString
        
        err := rows.Scan(
            &f.ID,
            &f.Name,
            &f.Email,
            &f.Rating,
            &comments,
            &f.CreatedAt,
            &f.UpdatedAt,
        )
        if err != nil {
            return nil, 0, err
        }
        
        if comments.Valid {
            f.Comments = comments.String
        }
        
        feedbacks = append(feedbacks, f)
    }
    
    return feedbacks, total, nil
}

// GetStats retrieves statistics about feedback
func (m *FeedbackModel) GetStats() (*StatsResponse, error) {
    // Get total count
    var total int
    countQuery := `SELECT COUNT(*) FROM feedback`
    err := m.DB.QueryRow(countQuery).Scan(&total)
    if err != nil {
        return nil, err
    }
    
    // Get average rating
    var avgRating float64
    avgQuery := `SELECT COALESCE(AVG(rating), 0) FROM feedback`
    err = m.DB.QueryRow(avgQuery).Scan(&avgRating)
    if err != nil {
        return nil, err
    }
    
    // Get rating distribution
    ratingQuery := `SELECT rating, COUNT(*) as count FROM feedback GROUP BY rating ORDER BY rating`
    rows, err := m.DB.Query(ratingQuery)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    ratingDistribution := make(map[int]int)
    for rows.Next() {
        var rating, count int
        err := rows.Scan(&rating, &count)
        if err != nil {
            return nil, err
        }
        ratingDistribution[rating] = count
    }
    
    stats := &StatsResponse{
        TotalFeedback:      total,
        AverageRating:      avgRating,
        RatingDistribution: ratingDistribution,
    }
    
    // Fill in missing ratings (1-5) with 0
    for i := 1; i <= 5; i++ {
        if _, exists := ratingDistribution[i]; !exists {
            ratingDistribution[i] = 0
        }
    }
    
    return stats, nil
}