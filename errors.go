package handlers

import (
    "database/sql"
    "errors"
    "net/http"
    "strings"
)

// Custom error types
var (
    ErrInvalidRequest    = errors.New("invalid request body")
    ErrNameRequired      = errors.New("name is required")
    ErrEmailRequired     = errors.New("email is required")
    ErrInvalidEmail      = errors.New("invalid email format")
    ErrInvalidRating     = errors.New("rating must be between 1 and 5")
    ErrFeedbackNotFound  = errors.New("feedback not found")
    ErrDatabaseError     = errors.New("database operation failed")
    ErrInvalidID         = errors.New("invalid ID format")
    ErrDuplicateEntry    = errors.New("feedback already exists")
    ErrForeignKeyViolation = errors.New("referenced record does not exist")
)

// HTTPStatusMapping returns the appropriate HTTP status code for each error
var HTTPStatusMapping = map[error]int{
    ErrInvalidRequest:    http.StatusBadRequest,
    ErrNameRequired:      http.StatusBadRequest,
    ErrEmailRequired:     http.StatusBadRequest,
    ErrInvalidEmail:      http.StatusBadRequest,
    ErrInvalidRating:     http.StatusBadRequest,
    ErrInvalidID:         http.StatusBadRequest,
    ErrFeedbackNotFound:  http.StatusNotFound,
    ErrDatabaseError:     http.StatusInternalServerError,
    ErrDuplicateEntry:    http.StatusConflict,
    ErrForeignKeyViolation: http.StatusBadRequest,
}

// GetHTTPStatus returns the HTTP status code for a given error
func GetHTTPStatus(err error) int {
    if status, exists := HTTPStatusMapping[err]; exists {
        return status
    }
    return http.StatusInternalServerError
}

// HandleDBError processes database errors and returns appropriate custom errors
func HandleDBError(err error) error {
    if err == nil {
        return nil
    }
    
    // Handle no rows found
    if err == sql.ErrNoRows {
        return ErrFeedbackNotFound
    }
    
    // Handle MySQL duplicate entry error (1062)
    if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "1062") {
        return ErrDuplicateEntry
    }
    
    // Handle MySQL foreign key constraint error (1452)
    if strings.Contains(err.Error(), "foreign key constraint") || strings.Contains(err.Error(), "1452") {
        return ErrForeignKeyViolation
    }
    
    // Handle connection errors
    if strings.Contains(err.Error(), "connection refused") || 
       strings.Contains(err.Error(), "deadlock") ||
       strings.Contains(err.Error(), "lock wait timeout") {
        return ErrDatabaseError
    }
    
    // Return generic database error for unknown issues
    return ErrDatabaseError
}

// ValidateFeedbackRequest validates the feedback input fields
func ValidateFeedbackRequest(name, email string, rating int) error {
    // Validate name
    if strings.TrimSpace(name) == "" {
        return ErrNameRequired
    }
    
    // Validate email
    if strings.TrimSpace(email) == "" {
        return ErrEmailRequired
    }
    
    // Basic email format validation
    if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
        return ErrInvalidEmail
    }
    
    // Validate email length (simple check)
    if len(email) < 5 || len(email) > 100 {
        return ErrInvalidEmail
    }
    
    // Validate rating
    if rating < 1 || rating > 5 {
        return ErrInvalidRating
    }
    
    return nil
}

// ValidateID validates that the ID is positive
func ValidateID(id int) error {
    if id <= 0 {
        return ErrInvalidID
    }
    return nil
}

// IsNotFoundError checks if the error is a not found error
func IsNotFoundError(err error) bool {
    return errors.Is(err, ErrFeedbackNotFound)
}

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool {
    validationErrors := []error{
        ErrNameRequired,
        ErrEmailRequired,
        ErrInvalidEmail,
        ErrInvalidRating,
        ErrInvalidID,
    }
    
    for _, validationErr := range validationErrors {
        if errors.Is(err, validationErr) {
            return true
        }
    }
    return false
}

// IsDatabaseError checks if the error is a database error
func IsDatabaseError(err error) bool {
    return errors.Is(err, ErrDatabaseError) || 
           errors.Is(err, ErrDuplicateEntry) ||
           errors.Is(err, ErrForeignKeyViolation)
}

// ErrorResponse represents the structure of error responses
type ErrorResponse struct {
    Error   string `json:"error"`
    Code    string `json:"code,omitempty"`
    Details string `json:"details,omitempty"`
}

// CreateErrorResponse creates a formatted error response
func CreateErrorResponse(err error, details string) ErrorResponse {
    response := ErrorResponse{
        Error: err.Error(),
    }
    
    // Add error codes for specific errors
    switch err {
    case ErrInvalidRequest:
        response.Code = "INVALID_REQUEST"
    case ErrNameRequired:
        response.Code = "NAME_REQUIRED"
    case ErrEmailRequired:
        response.Code = "EMAIL_REQUIRED"
    case ErrInvalidEmail:
        response.Code = "INVALID_EMAIL"
    case ErrInvalidRating:
        response.Code = "INVALID_RATING"
    case ErrFeedbackNotFound:
        response.Code = "NOT_FOUND"
    case ErrDatabaseError:
        response.Code = "DATABASE_ERROR"
    case ErrDuplicateEntry:
        response.Code = "DUPLICATE_ENTRY"
    default:
        response.Code = "UNKNOWN_ERROR"
    }
    
    if details != "" {
        response.Details = details
    }
    
    return response
}