package model

// Order сохраняет модель из предыдущего занятия.
type Order struct {
	ID       int
	Customer string
	Status   string
}

func ValidateStatus(status string) bool {
	switch status {
	case "new", "cancelled":
		return true
	default:
		return false
	}
}

type ListOptions struct {
	Limit  int
	Offset int
	Status string
}
