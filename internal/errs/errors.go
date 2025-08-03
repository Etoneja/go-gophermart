package errs

import "errors"

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrTokenExpired = errors.New("token expired")
var ErrUserExists = errors.New("user already exists")
var ErrOrderExists = errors.New("order already exists")
var ErrInsufficientFunds = errors.New("insufficient funds")
var ErrNotOnlyOneRowAffected = errors.New("zero or more than one row affected")
var ErrNoRows = errors.New("no rows")
