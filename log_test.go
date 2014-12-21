package spinlog

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"testing"
)

func TestSingleFile(t *testing.T) {
	cfg := *new(Config)
	cfg.Prefix = "yo"
	cfg.MaxCount = 1
	cfg.MaxSize = 10
	verifyLog(t, cfg, []string{"hey"}, []string{"hey"})
	verifyLog(t, cfg, []string{"0123456789"}, []string{"0123456789"})
	verifyLog(t, cfg, []string{"0123456789a"}, []string{"a"})
	verifyLog(t, cfg, []string{"01234", "56789a"}, []string{"a"})
}

func TestSingleRotate(t *testing.T) {
	cfg := *new(Config)
	cfg.Prefix = "yo"
	cfg.MaxCount = 2
	cfg.MaxSize = 10
	verifyLog(t, cfg, []string{"hey"}, []string{"hey"})
	verifyLog(t, cfg, []string{"0123456789"}, []string{"0123456789"})
	verifyLog(t, cfg, []string{"0123456789a"}, []string{"a", "0123456789"})
	verifyLog(t, cfg, []string{"01234", "56789a"}, []string{"a", "0123456789"})

	cfg.MaxSize = 2
	verifyLog(t, cfg, []string{"123", "456789"}, []string{"9", "78"})
}

func TestMultiRotate(t *testing.T) {
	cfg := *new(Config)
	cfg.Prefix = "hey"
	cfg.MaxSize = 3
	oldArray := []string{"fg", "cde", "9ab", "678", "345", "012"}
	input := []string{"01", "2345", "6789abc", "d", "e", "f", "g"}
	for i := 1; i < len(oldArray); i++ {
		cfg.MaxCount = i
		verifyLog(t, cfg, input, oldArray[0:i])
	}
}

func verifyLog(t *testing.T, c Config, writes []string, results []string) {
	dir, err := ioutil.TempDir("", "spinlog_test")
	if err != nil {
		t.Error("Failed to create temporary directory:", err)
		return
	}
	defer os.RemoveAll(dir)
	c.Directory = dir
	log, err := NewLog(c)
	if err != nil {
		t.Error("Failed to create log:", err)
		return
	}
	// Perform the writes
	for _, write := range writes {
		_, err := log.Write([]byte(write))
		if err != nil {
			t.Error("Write failed:", err)
			return
		}
	}
	if err := log.Close(); err != nil {
		t.Error("Close failed:", err)
	}
	// Check the results
	for i, expected := range results {
		filePath := path.Join(dir, c.Prefix + "." + strconv.Itoa(i))
		if content, err := ioutil.ReadFile(filePath); err != nil {
			t.Error("Failed to read file:", err)
		} else if !bytes.Equal(content, []byte(expected)) {
			t.Error("Unexpected content for file:", i)
		}
	}
	for i := 0; i < 100; i++ {
		badIndex := strconv.Itoa(i + len(results))
		filePath := path.Join(dir, c.Prefix + "." + badIndex)
		if _, err := ioutil.ReadFile(filePath); err == nil {
			t.Error("File should not exist at index:", i)
		}
	}
}
