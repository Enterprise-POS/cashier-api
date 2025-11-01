package exception

type PostgreSQLException struct {
	Code    string `json:"code"`
	Details string `json:"details"`
	Hint    string `json:"hint"`
	Message string `json:"message"`
}

func (e *PostgreSQLException) Error() string {
	return e.Message
}
