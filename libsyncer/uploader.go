package libsyncer

import (
	"io"
	"log"
	"net/http"
)

type Uploader struct {
	Volume Volume
}

func (u *Uploader) Upload(file FileID, peer PeerID, uploadURL string, done chan<- FileID) {
	log.Printf("Uploading file %s to %s\n", file, peer)

	if u.Volume.ID() != file.VolumeID {
		panic("Uploading invalid volume-id!")
	}

	reader, err := u.Volume.Read(file.Path)
	if err != nil {
		panic("Cannot read file: " + err.Error())
	}
	if closer, ok := reader.(io.Closer); ok {
		defer closer.Close()
	}

	req, err := http.NewRequest("PUT", uploadURL, reader)
	if err != nil {
		panic("Failed to build request: " + err.Error())
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic("Failed to upload " + file.String() + ": " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		done <- file
	}

	log.Println(file.String() + ": " + resp.Status)
}
