package postgres

import "fmt"

type AccountNotFound struct {
	Email string
}

func (m *AccountNotFound) Error() string {
	return fmt.Sprintf("account not found for email %s", m.Email)
}
