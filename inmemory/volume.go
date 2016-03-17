package inmemory

import (

	"time"
	"path/filepath"
)

type File struct {
	Path    string
	Name    string
	ModTime time.Time
	Size    uint64
	Dir     bool
}

type Volume struct {
	id    string
	Files map[string]File
	Size  uint64
}

func NewVolume(id string, size uint64) *Volume {
	return &Volume{
		id:    id,
		Files: make(map[string]File),
		Size:  size,
	}
}

func (v *Volume) ID() string {
	return v.id
}

func (v *Volume) AvailableBytes() uint64 {
	space := v.Size
	for _, file := range v.Files {
		space -= file.Size
	}
	return space
}

func (v *Volume) Walk(f filepath.WalkFunc) error {
	return nil
}
