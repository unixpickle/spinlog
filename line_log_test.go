package spinlog

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"testing"
)

func TestLineSingleFile(t *testing.T) {
	cfg := *new(LineConfig)
	cfg.MaxLineSize = 100
	cfg.MaxCount = 1
	cfg.MaxSize = 100
	verifyLineLog(t, cfg, []string{"hey\nthere\nbro\nyo!"},
		[]string{"hey\nthere\nbro\nyo!"})
	verifyLineLog(t, cfg, []string{"yo", " there\n", "what is up?", "\na"},
		[]string{"yo there\nwhat is up?\na"})
}

func TestLineSingleRotate(t *testing.T) {
	cfg := *new(LineConfig)
	cfg.MaxLineSize = 4
	cfg.MaxCount = 2
	cfg.MaxSize = 5
	verifyLineLog(t, cfg, []string{"hey\n", "yay\n"},
		[]string{"yay\n", "hey\n"})
	verifyLineLog(t, cfg, []string{"hey\nhey"}, []string{"hey", "hey\n"})
	verifyLineLog(t, cfg, []string{"a\nb\ny\n"}, []string{"y\n", "a\nb\n"})
	verifyLineLog(t, cfg, []string{"foo\n", "a", "\nb\nc\n"},
		[]string{"c\n", "a\nb\n"})
}

func TestLineOverflow(t *testing.T) {
	cfg := *new(LineConfig)
	cfg.MaxLineSize = 3
	cfg.MaxCount = 2
	cfg.MaxSize = 5
	verifyLineLog(t, cfg, []string{"fooba\n"}, []string{"ba\n", "foo"})
	verifyLineLog(t, cfg, []string{"foobar\n"}, []string{"bar\n", "foo"})
	verifyLineLog(t, cfg, []string{"foobarbaz\n"}, []string{"baz\n", "bar"})
	verifyLineLog(t, cfg, []string{"ab\naoeu\n"},
		[]string{"aoeu\n", "ab\n"})

	cfg.MaxCount = 3
	verifyLineLog(t, cfg, []string{"ab\naoeug\n"},
		[]string{"ug\n", "aoe", "ab\n"})
}

func verifyLineLog(t *testing.T, c LineConfig, writes []string,
	results []string) {
	c.LogDir.Prefix = "linelog"
	dir, err := ioutil.TempDir("", "spinlog_test")
	if err != nil {
		t.Error("Failed to create temporary directory:", err)
		return
	}
	defer os.RemoveAll(dir)
	c.LogDir.Directory = dir
	log, err := NewLineLog(c)
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
		filePath := path.Join(dir, c.LogDir.Prefix+"."+strconv.Itoa(i))
		if content, err := ioutil.ReadFile(filePath); err != nil {
			t.Error("Failed to read file:", err)
		} else if !bytes.Equal(content, []byte(expected)) {
			t.Error("Unexpected content for file:", i)
		}
	}
	for i := 0; i < 100; i++ {
		badIndex := strconv.Itoa(i + len(results))
		filePath := path.Join(dir, c.LogDir.Prefix+"."+badIndex)
		if _, err := ioutil.ReadFile(filePath); err == nil {
			t.Error("File should not exist at index:", i)
		}
	}
}
