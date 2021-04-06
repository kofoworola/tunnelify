package handler

import (
	"fmt"
	"io"
	"io/ioutil"
)

type ProxyHandler struct {
	reader io.Reader
}

func NewProxyHandler(reader io.Reader) *ProxyHandler {
	return &ProxyHandler{
		reader: reader,
	}
}

func (p *ProxyHandler) Handle() {
	fmt.Println("handling")
	dat, err := ioutil.ReadAll(p.reader)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(dat) + "hmm")
}
