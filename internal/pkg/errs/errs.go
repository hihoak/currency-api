package errs

import "fmt"

var (
	ErrUserAlreadyExists = fmt.Errorf("user already exists")
	ErrNotFound = fmt.Errorf("not found")

	// Database errors
	ErrConnectionFailed = fmt.Errorf("failed to connect to database")
	ErrPingFailed = fmt.Errorf("failed to ping database")
	ErrCloseConnectionFailed = fmt.Errorf("failed to close connection to database")
)
