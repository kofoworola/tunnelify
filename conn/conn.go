package conn

import (
	"bufio"
	"context"
	"fmt"
	"net"
)

type Connection struct {
	rw net.Conn
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{conn}
}

func (c *Connection) HandleConnection(ctx context.Context) {
	scanner := bufio.NewScanner(c.rw)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	fmt.Println("done getting connection req")
}
