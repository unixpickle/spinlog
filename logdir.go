package spinlog

import (
	"io"
	"os"
	"path"
	"sort"
	"strconv"
)

// LogDir stores a log directory and a file prefix.
type LogDir struct {
	Directory string `json:"directory"`
	Prefix    string `json:"prefix"`
}

// NewLogDir creates a new LogDir instance given a directory and prefix.
func NewLogDir(dir, prefix string) LogDir {
	return LogDir{dir, prefix}
}

// List reads the directory to find numerically labeled log files.
// For example, if the prefix is "foo" and the directory contains files called
// ["foo.0", "foo.24", "foobar", "foo.3"], this will return [0, 3, 24].
// The result is always sorted in ascending order.
// The result will never contain duplicate numbers.
// An error will be returned if the directory listing cannot be read.
func (l LogDir) List() ([]int, error) {
	// Open the directory
	f, err := os.Open(l.Directory)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Run Readdirnames again and again
	ids := make([]int, 0)
	for {
		names, err := f.Readdirnames(100)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		
		// Find identifiers in the names
		for _, name := range names {
			if id, ok := l.parseId(name); ok {
				ids = append(ids, id)
			}
		}
	}
	sort.Ints(ids)
	return ids, nil
}

// Rotate perfroms manual log rotation on a given directory.
// The max argument determines how many log files should be allowed to exist at
// once, including the 0 file which will not exist after this call.
// For example, if the prefix is "foo", max is 3, and the directory contains 
// files called ["foo.0", "foo.1", "foo.2", "foo.3", "foo.4", "foobar"], this
// will delete "foo.2", "foo.3", and "foo.4". It will rename "foo.1" to "foo.2"
// and "foo.0" to "foo.1".
// Usually, you do not need to use this method directly. Log and LineLog run it
// for you.
// An error will be returned if the directory cannot be read, if a file cannot
// be removed, or if a file cannot be renamed.
func (l LogDir) Rotate(max int) error {
	ids, err := l.List()
	if err != nil {
		return err
	}
	for i := len(ids) - 1; i >= 0; i-- {
		id := ids[i]
		filePath := l.FilePath(id)
		if id >= max-1 {
			if err := os.Remove(filePath); err != nil {
				return err
			}
		} else {
			newPath := l.FilePath(id+1)
			if err := os.Rename(filePath, newPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// FilePath returns the path to a log file with a given index.
func (l LogDir) FilePath(index int) string {
	basename := l.Prefix + "." + strconv.Itoa(index)
	return path.Join(l.Directory, basename)
}

func (l LogDir) parseId(name string) (int, bool) {
	prefixLen := len(l.Prefix)
	if len(name) < prefixLen+2 || name[0:prefixLen] != l.Prefix ||
		name[prefixLen] != '.' {
		return 0, false
	}
	num, err := strconv.Atoi(name[prefixLen+1:])
	if err != nil {
		return 0, false
	}
	return num, true
}
