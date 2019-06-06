package test_test

import (
	"io/ioutil"
	"path/filepath"
	"runtime"
	"testing"
)

// LoadBytes is helper to load file contents from testdata directory
func LoadBytes(t *testing.T, name string) []byte {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)

	path := filepath.Join(basepath, "testdata", name) // relative path
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

// LoadBytesRelative is helper to load file contents from testdata directory in test directory
func LoadBytesRelative(t *testing.T, path string) []byte {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}
