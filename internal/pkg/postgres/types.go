package postgres

import "fmt"

// TODO: support email and stripecustid. Or an arbitrary field?
type AccountNotFound struct {
	Email string
}

func (m *AccountNotFound) Error() string {
	return fmt.Sprintf("account not found")
}
