package handlers

import "errors"

var (
    ErrInvalidRequest = errors.New("invalid request")
    ErrNameRequired   = errors.New("name required")
    ErrEmailRequired  = errors.New("email required")
    ErrInvalidRating  = errors.New("rating must be 1-5")
    ErrNotFound       = errors.New("not found")
)

func GetHTTPStatus(err error) int {
    switch err {
    case ErrInvalidRequest, ErrNameRequired, ErrEmailRequired, ErrInvalidRating:
        return 400
    case ErrNotFound:
        return 404
    default:
        return 500
    }
}