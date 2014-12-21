package spinlog

import (
	"io"
	"os"
	"sync"
)

type Config struct {
	LogDir   LogDir `json:"log_dir"`
	MaxCount int    `json:"max_count"`
	MaxSize  int64  `json:"max_size"`
	SetPerm  bool   `json:"set_perm"`
	Perm     int    `json:"perm"`
	SetOwner bool   `json:"set_owner"`
	GID      int    `json:"gid"`
	UID      int    `json:"uid"`
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
	err := r.config.LogDir.Rotate(r.config.MaxCount)
	if err != nil {
		return err
	}

	// Open a new log file
	newPath := r.config.LogDir.FilePath(0)
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

func (l *Log) writeInternal(p []byte) (int, error) {
	// Make sure we are not closed
	if l.file == nil {
		return 0, io.ErrClosedPipe
	}

	// Write and rotate as many times as needed.
	written := 0
	for len(p) != 0 {
		toWrite, newP, err := l.splitBuffer(p)
		p = newP
		if err != nil {
			return written, err
		}
				
		w, err := l.file.Write(toWrite)
		written += w
		if err != nil {
			return written, err
		}
		
		if len(p) != 0 {
			if err := l.rotate(); err != nil {
				return written, err
			}
		}
	}

	return written, nil
}

func (r *Log) splitBuffer(p []byte) ([]byte, []byte, error) {
	remaining := int64(len(p))
	space, err := r.freeSpace()
	if err != nil {
		return nil, nil, err
	}
	if space >= remaining {
		return p, []byte{}, nil
	} else {
		return p[0:int(space)], p[int(space):], nil
	}
}
