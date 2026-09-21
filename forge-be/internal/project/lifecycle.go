package project

import "workspace/pkg/apperr"

var allowedTransitions = map[string][]string{
	StatusDraft:     {StatusActive, StatusOnHold},
	StatusActive:    {StatusOnHold, StatusCompleted},
	StatusOnHold:    {StatusActive, StatusArchived},
	StatusCompleted: {StatusArchived},
	StatusArchived:  {},
}

func CanTransition(from, to string) bool {
	if from == to {
		return true
	}
	for _, next := range allowedTransitions[from] {
		if next == to {
			return true
		}
	}
	return false
}

func ValidateTransition(from, to string) error {
	if CanTransition(from, to) {
		return nil
	}
	return apperr.ErrValidation.WithMessage("illegal project status transition from " + from + " to " + to)
}
