package modrinth

import "fmt"

type Error struct {
	Error       string `json:"error"`
	Description string `json:"description"`
}

func (e *Error) String() string {
	return fmt.Sprint(e.Error, e.Description)
}
