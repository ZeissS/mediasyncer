package mediasyncer

import (
	"io"
	"os"
	"path/filepath"

	"github.com/ricochet2200/go-disk-usage/du"
	"github.com/satori/go.uuid"
)

const (
	VolumeIDFile = ".mediasyncer-volume-id"
)

type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

// A Volume represents a folder on disk which can contain a lot of (big and small) files.
type Volume struct {
	ID   string
	Path string
}

func OpenVolume(volumePath string) *Volume {
	if volumePath == "" {
		panic("Volume path must not be empty.")
	}

	var id string
	filepath := filepath.Join(volumePath, VolumeIDFile)
	fp, err := os.Open(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			fp, err = os.Create(filepath)
			if err != nil {
				panic(err.Error())
			}
			id = uuid.NewV4().String()

			_, err = fp.WriteString(id)
			if err != nil {
				panic(err.Error())
			}
			fp.Sync()

		} else {
			panic(err.Error())
		}
	} else {
		buffer := make([]byte, 256)
		n, err := fp.Read(buffer)
		if err != nil {
			panic(err.Error())
		}

		id = string(buffer[:n])
	}

	if id == "" {
		panic("Read empty volume-id from path " + volumePath)
	}

	if err = fp.Close(); err != nil {
		panic(err.Error())
	}

	return &Volume{
		ID:   id,
		Path: volumePath,
	}
}

func (v *Volume) AvailableBytes() uint64 {
	return du.NewDiskUsage(v.Path).Available()
}

func (v *Volume) Walk(f filepath.WalkFunc) error {
	//fmt.Println("# " + v.Path)
	return filepath.Walk(v.Path, func(fullpath string, info os.FileInfo, err error) error {
		//	fmt.Println("> " + fullpath + "\t" + info.Name())
		if info.IsDir() {
			return nil
		}

		base := filepath.Base(fullpath)
		if base == VolumeIDFile {
			return nil
		}

		relPath, err := filepath.Rel(v.Path, fullpath)
		if err != nil {
			return err
		}

		//	fmt.Println("? " + relPath)
		return f(relPath, info, err)
	})
}

func (v *Volume) Stat(path string) (os.FileInfo, error) {
	fp := filepath.Join(v.Path, path)
	return os.Stat(fp)
}

func (v *Volume) Read(path string) (ReadSeekCloser, error) {
	fp := filepath.Join(v.Path, path)
	return os.Open(fp)
}

func (v *Volume) Write(path string) (io.WriteCloser, error) {
	fp := filepath.Join(v.Path, path)
	return os.Create(fp)
}

func (v *Volume) Delete(path string) error {
	fp := filepath.Join(v.Path, path)

	return os.Remove(fp)
}
