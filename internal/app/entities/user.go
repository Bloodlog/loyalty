package entities

import "database/sql"

type User struct {
	Balance  sql.NullFloat64
	Login    string
	Password string
	ID       int
}
