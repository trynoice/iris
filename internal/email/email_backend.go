package email

type EmailBackend interface {
	Send(e *Email) error
}
