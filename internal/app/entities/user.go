package entities

import "database/sql"

type User struct {
	Login    string
	Password string
	Balance  sql.NullFloat64
	ID       int
}
