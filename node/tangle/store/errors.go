package tangle

import (
	"errors"
	"strconv"
)

var (
	// ErrWeightTooLow is returned when the weight does not exceed MinimumWeight
	ErrWeightTooLow = errors.New("Weight too low. Has to be > " + strconv.Itoa(MinimumWeight))
	// ErrNotValidating is returned when the site does not validate any current tip
	ErrNotValidating = errors.New("Site does not validate any current tip")
	// ErrTooFewValidations is returned when the site does not validate enough sites
	ErrTooFewValidations = errors.New("Site does not validate enough sites")
)
