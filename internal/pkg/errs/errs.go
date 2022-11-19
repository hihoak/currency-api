package errs

import "fmt"

var (
	ErrUserAlreadyExists = fmt.Errorf("user already exists")
	ErrFailedToGetUnnesesaryInformation = fmt.Errorf("main logic is done, but can't get additional info")

	// Database errors
	ErrConnectionFailed = fmt.Errorf("failed to connect to database")
	ErrPingFailed = fmt.Errorf("failed to ping database")
	ErrCloseConnectionFailed = fmt.Errorf("failed to close connection to database")
)
