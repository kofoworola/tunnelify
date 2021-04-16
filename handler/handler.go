package handler

type CloseFunc func() error

type ConnectionHandler interface {
	// TODO think of using context to handle shutdown
	Handle()
}
