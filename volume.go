package mediasyncer

import (
	"io"
	"os"
	"path/filepath"
)

type Volume interface {
	ID() string
	AvailableBytes() uint64
	Walk(f filepath.WalkFunc) error

	Stat(path string) (os.FileInfo, error)
	Read(path string) (io.ReadSeeker, error)
	Write(path string) (io.WriteCloser, error)
	Delete(path string) error
}
