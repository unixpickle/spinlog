package spinlog

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"testing"
)

func TestLogDirList(t *testing.T) {
	dir, err := createTempDir()
	if err != nil {
		t.Fatal("Failed to create temporary directory:", err)
	}
	defer os.RemoveAll(dir)
	list, err := NewLogDir(dir, "log").List()
	if err != nil {
		t.Fatal("Failed to list:", err)
	}
	idx := 0
	for i := 0; i < 1000; i++ {
		if (i + 1) % 33 == 0 {
			continue
		}
		if list[idx] != i {
			t.Error("Unexpected element in list", list[idx], "expected", i)
		}
		idx++
	}
	if idx != len(list) {
		t.Error("List had unexpected length.")
	}
}

func TestLogDirRotate(t *testing.T) {
	dir, err := createTempDir()
	if err != nil {
		t.Fatal("Failed to create temporary directory:", err)
		return
	}
	defer os.RemoveAll(dir)
	ld := NewLogDir(dir, "log")
	if err := ld.Rotate(1000); err != nil {
		t.Error("Rotate failed:", err)
		return
	}
	list, err := ld.List()
	if err != nil {
		t.Error("Failed to list:", err)
		return
	}
	idx := 0
	for i := 1; i < 1000; i++ {
		if i % 33 == 0 {
			continue
		}
		if list[idx] != i {
			t.Error("Unexpected element in list", list[idx], "expected", i)
		}
		idx++
	}
	if idx != len(list) {
		t.Error("List had unexpected length.")
	}
}

func createTempDir() (string, error) {
	dir, err := ioutil.TempDir("", "spinlog_test")
	if err != nil {
		return "", err
	}
	for i := 0; i < 1000; i++ {
		if (i + 1) % 33 == 0 {
			continue
		}
		name := path.Join(dir, "log." + strconv.Itoa(i))
		f, err := os.Create(name)
		if err != nil {
			os.RemoveAll(dir)
			return "", err
		}
		f.Close()
	}
	return dir, nil
}
