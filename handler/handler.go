package handler

type CloseFunc func() error

type ConnectionHandler interface {
	Handle()
}
