package spinlog

import "bytes"

type LineRotatorConfig struct {
	RotatorConfig
	MaxLineSize int `json:"max_line_size"`
}

type LineRotator struct {
	*Rotator
	lineSize int
	line     *bytes.Buffer
}

func NewLineRotator(config LineRotatorConfig) (*LineRotator, error) {
	if int64(config.MaxLineSize) > config.MaxSize {
		return nil, ErrBadConfig
	}
	rot, err := NewRotator(config.RotatorConfig)
	if err != nil {
		return nil, err
	}
	return &LineRotator{rot, config.MaxLineSize, new(bytes.Buffer)}, nil
}

func (l *LineRotator) Close() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	// Return the flush error only if some other error doesn't occur
	// simultaneously.
	var flushErr error
	if l.line.Len() > 0 {
		flushErr = l.flushLine()
	}
	
	if err := l.closeInternal(); err != nil {
		return err
	}
	return flushErr
}

func (l *LineRotator) Write(p []byte) (int, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	for i, b := range p {
		l.line.WriteByte(b)
		if l.line.Len() >= l.lineSize || b == '\n' {
			// Flush the line
			if err := l.flushLine(); err != nil {
				return i, err
			}
		}
	}
	return len(p), nil
}

func (l *LineRotator) flushLine() error {
	rem, err := l.freeSpace()
	if err != nil {
		return err
	}
	if int64(l.line.Len()) > rem {
		if err := l.rotate(); err != nil {
			return err
		}
	}
	_, err = l.writeInternal(l.line.Bytes())
	l.line = new(bytes.Buffer)
	return err
}
