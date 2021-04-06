package tunnelify

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestReadWrapperReadWrite(t *testing.T) {
	const (
		stringData  = "This is some random test. please God, let it work"
		writtenData = "Hello World\n"
	)
	rw := NewConnectionReader([]byte(writtenData), bytes.NewBuffer([]byte(stringData)), ioutil.Discard)
	dat, err := ioutil.ReadAll(rw)
	if err != nil {
		t.Fatalf("error reading all data: %v", err)
	}
	if string(dat) != writtenData+stringData {
		t.Fatalf("expected: '%s' got:'%s'", writtenData+stringData, string(dat))
	}

}
