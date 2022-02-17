package downloads

import (
	"errors"
	"log"
)

type incompleteFile struct {
	ID           int         `json:"id"`
	TargetLength int         `json:"len"`
	ChunkSize    int         `json:"c"`
	Splits       []fileSplit `json:"s"`
}

type fileSplit struct {
	ChunksComplete int `json:"c"`
}

func (c *Client) download(d Download) {
	if d.Status >= Downloading {
		return
	}
	if d.transferInfo.Url == "" {
		// TODO: Attempt to load .incomplete file
		return
	}
	log.Printf("Starting download for %d, %s\n", d.ID, d.transferInfo.Url)
	c.updateStatus(d.ID, Downloading)
	c.mutex.Lock()
	c.slots--
	c.mutex.Unlock()
	header, err := Head(d.transferInfo.Url)
	if err != nil {
		log.Println(err)
		c.updateStatus(d.ID, Error)
		return
	}
	if header.ContentLength <= 0 {
		log.Println(errors.New("file was empty"))
		c.updateStatus(d.ID, Error)
		return
	}
	log.Println("Len: ", header.ContentLength)
	progress := incompleteFile{
		ID:           d.ID,
		TargetLength: header.ContentLength,
		ChunkSize:    95, // mb
		Splits:       make([]fileSplit, c.opt.Splits),
	}
	// TODO: Create file and fill to total length
	// TODO: Check if file length has changed if a .incomplete file was loaded
	// TODO: Check hash of file?
	for i, val := range progress.Splits {
		log.Printf("split index: %d, complete: %d\n", i, val.ChunksComplete)
		go progress.split(val)
	}
}

func (f *incompleteFile) split(s fileSplit) {
	// TODO
}
