package chunk

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func TestFileChunk(t *testing.T) {
	f := "./test_chunk"
	c, err := OpenFileChunk(f, 16)
	if err != nil {
		t.Fatalf("OpenFileChunk() returns err=%v", err)
	}
	defer os.Remove(f)

	if c.Checksum() != 0 {
		t.Errorf("Size() is mismatch: got=0x%016x, want=0x%016x", c.Checksum(), 0)
	}

	in := []byte("Hello")
	if n, err := c.Write(in); err != nil {
		t.Errorf("Write() returns err=%v", err)
	} else if n != 5 {
		t.Errorf("Write() returns wrong n: got=%d, want=%d", n, 5)
	}
	if c.Size() != 5 {
		t.Errorf("Size() is mismatch: got=%d, want=%d", c.Size(), 5)
	}
	if c.Checksum() != 0x51cf5c3bc87bacc8 {
		t.Errorf("Size() is mismatch: got=0x%016x, want=0x%016x", c.Checksum(), 0x51cf5c3bc87bacc8)
	}

	if n, err := c.Seek(0, io.SeekStart); err != nil {
		t.Errorf("Seek() returns err=%v", err)
	} else if n != 0 {
		t.Errorf("Seek() returns wrong n: got=%d, want=%d", n, 0)
	}

	b, err := ioutil.ReadAll(c)
	if err != nil {
		t.Errorf("Read() returns err=%v", err)
	}
	if !bytes.Equal(b, in) {
		t.Errorf("Read() is mismatch: got=%v, want=%v", b, in)
	}

	// Test rewrite
	in = []byte(" World!!")
	if n, err := c.Write(in); err != nil {
		t.Errorf("Write() returns err=%v", err)
	} else if n != len(in) {
		t.Errorf("Write() returns wrong n: got=%d, want=%d", n, len(in))
	}
	if c.Size() != 13 {
		t.Errorf("Size() is mismatch: got=%d, want=%d", c.Size(), 13)
	}
	if c.Checksum() != 0xf5a8a397b60da2e1 {
		t.Errorf("Size() is mismatch: got=0x%016x, want=0x%016x", c.Checksum(), uint64(0xf5a8a397b60da2e1))
	}

	if n, err := c.Seek(0, io.SeekStart); err != nil {
		t.Errorf("Seek() returns err=%v", err)
	} else if n != 0 {
		t.Errorf("Seek() returns wrong n: got=%d, want=%d", n, 0)
	}
	in = []byte("aaaaaaaaaaaaaaaa")
	if n, err := c.Write(in); err != nil {
		t.Errorf("Write() returns err=%v", err)
	} else if n != len(in) {
		t.Errorf("Write() returns wrong n: got=%d, want=%d", n, len(in))
	}
	if n, err := c.Write([]byte("a")); err != io.EOF {
		t.Errorf("Write() returns not io.EOF: err=%v", err)
	} else if n != 0 {
		t.Errorf("Write() returns wrong n: got=%d, want=%d", n, 0)
	}

	if err := c.Close(); err != nil {
		t.Errorf("Close() returns err=%v", err)
	}
}
