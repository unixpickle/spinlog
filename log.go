package spinlog

import (
	"io"
	"os"
	pathlib "path"
	"sort"
	"strconv"
	"sync"
)

type Config struct {
	Directory string `json:"directory"`
	Prefix    string `json:"prefix"`
	MaxCount  int    `json:"max_count"`
	MaxSize   int64  `json:"max_size"`
	SetPerm   bool   `json:"set_perm"`
	Perm      int    `json:"perm"`
	SetOwner  bool   `json:"set_owner"`
	GID       int    `json:"gid"`
	UID       int    `json:"uid"`
}

type Log struct {
	mutex  sync.Mutex
	config Config
	file   *os.File
}

func NewLog(c Config) (*Log, error) {
	if c.MaxCount < 1 || c.MaxSize < 1 {
		return nil, ErrBadConfig
	}
	res := &Log{sync.Mutex{}, c, nil}
	if err := res.rotate(); err != nil {
		return nil, err
	}
	return res, nil
}

func (r *Log) Close() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.closeInternal()
}

func (r *Log) Write(p []byte) (int, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.writeInternal(p)
}

func (r *Log) closeInternal() error {
	// If the file is nil, we are closed
	if r.file == nil {
		return io.ErrClosedPipe
	}

	// Close the file and set it to nil
	res := r.file.Close()
	r.file = nil
	return res
}

func (r *Log) freeSpace() (int64, error) {
	offset, err := r.file.Seek(0, 1)
	if err != nil {
		return 0, err
	}
	if offset > r.config.MaxSize {
		return 0, nil
	}
	return r.config.MaxSize - offset, nil
}

func (r *Log) rotate() error {
	// Make sure we don't have the latest file open
	if r.file != nil {
		r.file.Close()
		r.file = nil
	}

	// Attempt to rotate files
	err := rotateFiles(r.config.Directory, r.config.Prefix,
		r.config.MaxCount)
	if err != nil {
		return err
	}

	// Open a new log file
	newPath := logFilePath(r.config.Directory, r.config.Prefix, 0)
	flags := os.O_RDWR | os.O_CREATE | os.O_EXCL
	perm := 0600
	if r.config.SetPerm {
		perm = r.config.Perm
	}
	f, err := os.OpenFile(newPath, flags, os.FileMode(perm))
	if err != nil {
		return err
	}

	// Set the owner if specified
	if r.config.SetOwner {
		if err := f.Chown(r.config.UID, r.config.GID); err != nil {
			f.Close()
			return err
		}
	}

	r.file = f
	return nil
}

func (r *Log) writeInternal(p []byte) (int, error) {
	// Make sure we are not closed
	if r.file == nil {
		return 0, io.ErrClosedPipe
	}

	// Split up the bytes as necessary in order to avoid overflowing a log file.
	written := 0
	for {
		remaining := int64(len(p))
		space, err := r.freeSpace()
		if err != nil {
			return written, err
		}

		if space >= remaining {
			// There is enough space to write the entire thing.
			w, err := r.file.Write(p)
			written += w
			if err != nil {
				return written, err
			}
			break
		}

		// Split up the data
		w, err := r.file.Write(p[0:int(space)])
		written += w
		if err != nil {
			return written, err
		}
		if err := r.rotate(); err != nil {
			return written, err
		}
		p = p[int(space):]
	}

	return written, nil
}

func listIdentifiers(dir, prefix string) ([]int, error) {
	// Open the directory
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Run Readdirnames again and again
	ids := make([]int, 0)
	prefixLen := len(prefix)
	for {
		names, err := f.Readdirnames(100)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		// Find identifiers in the names
		for _, name := range names {
			if len(name) < prefixLen+2 || name[0:prefixLen] != prefix ||
				name[prefixLen] != '.' {
				continue
			}
			num, err := strconv.Atoi(name[prefixLen+1:])
			if err != nil {
				continue
			}
			ids = append(ids, num)
		}
	}
	sort.Ints(ids)
	return ids, nil
}

func logFilePath(dir, prefix string, id int) string {
	basename := prefix + "." + strconv.Itoa(id)
	return pathlib.Join(dir, basename)
}

func rotateFiles(dir, prefix string, max int) error {
	ids, err := listIdentifiers(dir, prefix)
	if err != nil {
		return err
	}
	for i := len(ids) - 1; i >= 0; i-- {
		id := ids[i]
		filePath := logFilePath(dir, prefix, id)
		if id >= max-1 {
			if err := os.Remove(filePath); err != nil {
				return err
			}
		} else {
			newPath := logFilePath(dir, prefix, id+1)
			if err := os.Rename(filePath, newPath); err != nil {
				return err
			}
		}
	}
	return nil
}
