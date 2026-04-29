package handlers

import (
    "crypto/rand"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "net/http"
    "regexp"
    "strings"
    "time"
)

// Helper function to generate a random ID (useful for request tracing)
func GenerateRequestID() string {
    bytes := make([]byte, 8)
    if _, err := rand.Read(bytes); err != nil {
        return fmt.Sprintf("%d", time.Now().UnixNano())
    }
    return hex.EncodeToString(bytes)
}

// Helper function to sanitize input strings (prevent XSS)
func SanitizeString(input string) string {
    // Remove any HTML tags
    re := regexp.MustCompile(`<[^>]*>`)
    sanitized := re.ReplaceAllString(input, "")
    
    // Trim whitespace
    sanitized = strings.TrimSpace(sanitized)
    
    // Limit length
    if len(sanitized) > 500 {
        sanitized = sanitized[:500]
    }
    
    return sanitized
}

// Helper function to validate email format more strictly
func IsValidEmail(email string) bool {
    // Basic email regex pattern
    pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
    re := regexp.MustCompile(pattern)
    return re.MatchString(email)
}

// Helper function to parse and validate date ranges
func ParseDateRange(startDate, endDate string) (time.Time, time.Time, error) {
    var start, end time.Time
    var err error
    
    if startDate != "" {
        start, err = time.Parse("2006-01-02", startDate)
        if err != nil {
            return time.Time{}, time.Time{}, fmt.Errorf("invalid start date format: %v", err)
        }
    } else {
        start = time.Now().AddDate(-1, 0, 0) // Default to 1 year ago
    }
    
    if endDate != "" {
        end, err = time.Parse("2006-01-02", endDate)
        if err != nil {
            return time.Time{}, time.Time{}, fmt.Errorf("invalid end date format: %v", err)
        }
    } else {
        end = time.Now()
    }
    
    if start.After(end) {
        return time.Time{}, time.Time{}, fmt.Errorf("start date cannot be after end date")
    }
    
    return start, end, nil
}

// Helper function to add CORS headers
func EnableCORS(w http.ResponseWriter) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
    w.Header().Set("Access-Control-Max-Age", "3600")
}

// Helper function to handle OPTIONS requests for CORS
func HandleCORS(w http.ResponseWriter, r *http.Request) bool {
    if r.Method == "OPTIONS" {
        EnableCORS(w)
        w.WriteHeader(http.StatusOK)
        return true
    }
    EnableCORS(w)
    return false
}

// Helper function to log requests (for debugging)
func LogRequest(r *http.Request, status int, duration time.Duration) {
    fmt.Printf("[%s] %s %s - Status: %d - Duration: %v\n",
        time.Now().Format("2006-01-02 15:04:05"),
        r.Method,
        r.URL.Path,
        status,
        duration,
    )
}

// Helper function to extract query parameters with defaults
func GetQueryParam(r *http.Request, key, defaultValue string) string {
    value := r.URL.Query().Get(key)
    if value == "" {
        return defaultValue
    }
    return value
}

// Helper function to get pagination parameters
func GetPaginationParams(r *http.Request) (limit, offset int) {
    limitStr := GetQueryParam(r, "limit", "10")
    offsetStr := GetQueryParam(r, "offset", "0")
    
    limit, _ = strconv.Atoi(limitStr)
    offset, _ = strconv.Atoi(offsetStr)
    
    // Validate limits
    if limit <= 0 {
        limit = 10
    }
    if limit > 100 {
        limit = 100
    }
    if offset < 0 {
        offset = 0
    }
    
    return limit, offset
}

// Helper function to build paginated response
func BuildPaginatedResponse(data interface{}, total, limit, offset int) map[string]interface{} {
    return map[string]interface{}{
        "data":       data,
        "total":      total,
        "limit":      limit,
        "offset":     offset,
        "next":       offset+limit < total,
        "previous":   offset > 0,
    }
}

// Helper function to mask email for privacy
func MaskEmail(email string) string {
    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return email
    }
    
    username := parts[0]
    domain := parts[1]
    
    if len(username) <= 2 {
        return strings.Repeat("*", len(username)) + "@" + domain
    }
    
    maskedUsername := username[:2] + strings.Repeat("*", len(username)-2)
    return maskedUsername + "@" + domain
}

// Helper function to validate JSON content type
func ValidateContentType(r *http.Request) error {
    contentType := r.Header.Get("Content-Type")
    if contentType != "application/json" {
        return ErrInvalidRequest
    }
    return nil
}

// Helper function to write a simple text response
func WriteText(w http.ResponseWriter, status int, message string) {
    w.Header().Set("Content-Type", "text/plain")
    w.WriteHeader(status)
    w.Write([]byte(message))
}

// Helper function to write XML response (if needed)
func WriteXML(w http.ResponseWriter, status int, xmlData string) {
    w.Header().Set("Content-Type", "application/xml")
    w.WriteHeader(status)
    w.Write([]byte(xmlData))
}

// Helper function to create a map for JSON responses quickly
func JSONMap(keyValues ...interface{}) map[string]interface{} {
    if len(keyValues)%2 != 0 {
        return map[string]interface{}{"error": "invalid number of arguments"}
    }
    
    result := make(map[string]interface{})
    for i := 0; i < len(keyValues); i += 2 {
        key, ok := keyValues[i].(string)
        if !ok {
            continue
        }
        result[key] = keyValues[i+1]
    }
    return result
}

// Helper function to check if a string is empty or only whitespace
func IsEmptyOrWhitespace(str string) bool {
    return strings.TrimSpace(str) == ""
}

// Helper function to truncate long strings
func TruncateString(str string, maxLength int) string {
    if len(str) <= maxLength {
        return str
    }
    return str[:maxLength] + "..."
}

// Helper function to convert slice of interfaces to slice of strings
func ToStringSlice(items []interface{}) []string {
    result := make([]string, len(items))
    for i, item := range items {
        result[i] = fmt.Sprintf("%v", item)
    }
    return result
}

// Helper function to check if a value exists in a slice
func Contains(slice []string, value string) bool {
    for _, item := range slice {
        if item == value {
            return true
        }
    }
    return false
}

// Helper function to get client IP address from request
func GetClientIP(r *http.Request) string {
    // Check for X-Forwarded-For header
    xff := r.Header.Get("X-Forwarded-For")
    if xff != "" {
        ips := strings.Split(xff, ",")
        return strings.TrimSpace(ips[0])
    }
    
    // Check for X-Real-IP header
    xri := r.Header.Get("X-Real-IP")
    if xri != "" {
        return xri
    }
    
    // Fall back to RemoteAddr
    ip := r.RemoteAddr
    if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
        ip = ip[:colonIndex]
    }
    
    return ip
}

// Helper function to rate limit based on IP (simple implementation)
type RateLimiter struct {
    requests map[string][]time.Time
    limit    int
    window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
    return &RateLimiter{
        requests: make(map[string][]time.Time),
        limit:    limit,
        window:   window,
    }
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow(ip string) bool {
    now := time.Now()
    
    // Clean up old requests
    if requests, exists := rl.requests[ip]; exists {
        var validRequests []time.Time
        for _, reqTime := range requests {
            if now.Sub(reqTime) < rl.window {
                validRequests = append(validRequests, reqTime)
            }
        }
        rl.requests[ip] = validRequests
    }
    
    // Check limit
    if len(rl.requests[ip]) >= rl.limit {
        return false
    }
    
    // Add current request
    rl.requests[ip] = append(rl.requests[ip], now)
    return true
}

// Global rate limiter instance (optional)
var DefaultRateLimiter = NewRateLimiter(100, time.Minute)

// Helper function to check rate limit
func CheckRateLimit(r *http.Request) bool {
    ip := GetClientIP(r)
    return DefaultRateLimiter.Allow(ip)
}

// Helper function to parse JSON safely with size limit
func SafeReadJSON(r *http.Request, dst interface{}, maxBytes int64) error {
    // Limit request body size
    r.Body = http.MaxBytesReader(nil, r.Body, maxBytes)
    
    decoder := json.NewDecoder(r.Body)
    decoder.DisallowUnknownFields() // Reject unknown fields
    
    err := decoder.Decode(dst)
    if err != nil {
        return fmt.Errorf("failed to parse JSON: %v", err)
    }
    
    // Check for extra data after JSON
    if decoder.More() {
        return fmt.Errorf("extra data after JSON object")
    }
    
    return nil
}