package tunnelify

import (
	"io/ioutil"
	"strings"
	"testing"
)

func TestByteReader(t *testing.T) {
	const (
		firstString  = "Hello world, please work"
		secondString = "Please"
	)
	reader := byteReader{
		s:        make([]byte, 0),
		i:        0,
		prevRune: -1,
	}
	reader.appendToData([]byte(firstString))
	reader.appendToData([]byte(secondString))

	gotten, err := ioutil.ReadAll(&reader)
	if err != nil {
		t.Fatalf("error reading reader: %v", err)
	}

	if string(gotten) != firstString+secondString {
		t.Fatalf("expected: %s\n got:%s\n", firstString+secondString, string(gotten))
	}
}

func TestReadWrapperReadWrite(t *testing.T) {
	const (
		stringData  = "This is some random test. please God, let it work"
		writtenData = "Hello World\n"
	)
	t.Run("TestCompleteSequentialRead", func(t *testing.T) {
		rw := NewReadWrapper(strings.NewReader(stringData))
		rw.Write([]byte(writtenData))
		dat, err := ioutil.ReadAll(rw)
		if err != nil {
			t.Fatalf("error reading all data: %v", err)
		}
		if string(dat) != writtenData+stringData {
			t.Fatalf("expected: %s\n got: %s", writtenData+stringData, string(dat))
		}
	})

	t.Run("TestPartialSquentialRead", func(t *testing.T) {
		rw := NewReadWrapper(strings.NewReader(stringData))
		rw.Write([]byte(writtenData))
		dat, err := ioutil.ReadAll(rw)
		if err != nil {
			t.Fatalf("error reading all data: %v", err)
		}
		if string(dat) != writtenData+stringData {
			t.Fatalf("expected: '%s' got:'%s'", writtenData+stringData, string(dat))
		}

	})

}
