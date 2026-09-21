package task

import "workspace/pkg/apperr"

var allowedTransitions = map[string][]string{
	StatusBacklog:    {StatusTodo, StatusBlocked},
	StatusTodo:       {StatusBacklog, StatusInProgress, StatusBlocked},
	StatusInProgress: {StatusTodo, StatusInReview, StatusBlocked, StatusDone},
	StatusInReview:   {StatusInProgress, StatusDone, StatusBlocked},
	StatusDone:       {StatusInReview},
	StatusBlocked:    {StatusBacklog, StatusTodo, StatusInProgress},
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
	return apperr.ErrValidation.WithMessage("illegal task status transition from " + from + " to " + to)
}

func IsStartStatus(status string) bool {
	switch status {
	case StatusTodo, StatusInProgress, StatusInReview:
		return true
	default:
		return false
	}
}
