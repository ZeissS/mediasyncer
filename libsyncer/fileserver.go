package libsyncer

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
)

type FileServerConfig struct {
	Addr string
	Port int
}

type FileServer struct {
	FileServerConfig

	Volume Volume

	l net.Listener
}

func NewFileServer(cfg FileServerConfig, vol Volume) *FileServer {
	return &FileServer{
		FileServerConfig: cfg,
		Volume:           vol,
	}
}

func (fs *FileServer) Serve() {
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", fs.Port))
	if err != nil {
		panic(err.Error())
	}
	fs.l = l

	if err := http.Serve(fs.l, fs); err != nil {
		log.Println("ERROR in http.Serve: " + err.Error())
	}
}

func (fs *FileServer) Close() {
	if err := fs.l.Close(); err != nil {
		panic(err.Error())
	}
}

// CreateUploadURL returns an URL that can be used to PUT the given file.
// The URL may be signed or have any number of query parameter.
// A client performing the upload MUST NOT modify this URL.
func (fs *FileServer) CreateUploadURL(file FileID) (string, error) {
	if file.VolumeID != fs.Volume.ID() {
		panic("Invalid volume id")
	}

	// TODO: End signature
	// TODO: End expire date
	return fmt.Sprintf("http://%s:%d/%s", fs.Addr, fs.Port, file.Path), nil
}

// HTTP Handler Implementation

func (fs *FileServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	if req.Method == "HEAD" || req.Method == "GET" {
		filepath := req.RequestURI
		stats, err := fs.Volume.Stat(filepath)
		if err != nil {
			if os.IsNotExist(err) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		file, err := fs.Volume.Read(filepath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.ServeContent(w, req, req.RequestURI, stats.ModTime(), file)
	} else if req.Method == "PUT" {
		path := req.RequestURI

		file := FileID{
			VolumeID: fs.Volume.ID(),
			Path:     path,
		}
		size := req.Header.Get("Content-Length")

		log.Println("Receiving upload for " + file.String() + " (size=" + size + ")")

		// We expect a does-not-exist error here.
		// no error => File exists => forbidden
		// does-not-exist => No file there => OK, go on
		// other error => internal server error
		_, err := fs.Volume.Stat(path)
		if err == nil {
			w.WriteHeader(http.StatusForbidden)
			return
		} else if !os.IsNotExist(err) {
			log.Println("ERROR Stat(): " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		writer, err := fs.Volume.Write(path)
		if err != nil {
			log.Println("ERROR Write(): " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer writer.Close()

		if _, err := io.Copy(writer, req.Body); err != nil {
			log.Println("ERROR: Failed to upload file: " + err.Error())
		}
		w.WriteHeader(http.StatusCreated)
		log.Printf("Upload of %v succeeded.\n", file)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}
