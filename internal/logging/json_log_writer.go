package logging

import (
	"io"
	"os"
	"sync"
)

type JSONLogWriter struct {
	file *os.File
	mu   sync.Mutex
}

func NewJSONLogWriter(file *os.File) *JSONLogWriter {
	return &JSONLogWriter{file: file}
}

func (w *JSONLogWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	stat, err := w.file.Stat()
	if err != nil {
		return 0, err
	}

	if stat.Size() == 0 {
		// If the file is empty, write the opening bracket
		if _, err := w.file.Write([]byte("[\n")); err != nil {
			return 0, err
		}
	} else {
		// If the file is not empty, seek to before the last byte (the closing ']')
		// and write a comma to separate the new log from the previous one
		if _, err := w.file.Seek(-2, io.SeekEnd); err != nil {
			return 0, err
		}
		if _, err := w.file.Write([]byte(",\n")); err != nil {
			return 0, err
		}
	}

	// Write the log message
	n, err = w.file.Write(p)
	if err != nil {
		return n, err
	}

	// Write the closing bracket
	if _, err := w.file.Write([]byte("\n]")); err != nil {
		return n, err
	}

	return n, nil
}
