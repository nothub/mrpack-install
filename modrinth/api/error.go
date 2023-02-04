package api

import "fmt"

type Error struct {
	Name        string `json:"error"`
	Description string `json:"description"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s (%s)", e.Description, e.Name)
}
