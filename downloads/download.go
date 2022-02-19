package downloads

import (
	"errors"
	"log"
	"os"
)

type incompleteFile struct {
	ID           int         `json:"id"`
	TargetLength int         `json:"len"`
	ChunkSize    int         `json:"c"`
	Splits       []fileSplit `json:"s"`
	handle       *os.File
}

type fileSplit struct {
	ChunksComplete int `json:"c"`
}

func (c *Client) download(d Download) {
	if d.Status >= Downloading {
		return
	}
	new := d.transferInfo.Url != ""
	if !new {
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
	incompleteFile := incompleteFile{
		ID:           d.ID,
		TargetLength: header.ContentLength,
		ChunkSize:    30 * 1e+6, // 30 mb
		Splits:       make([]fileSplit, c.opt.Splits),
	}
	handle, err := os.OpenFile(d.transferInfo.Path+".incomplete", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Println(err)
		c.updateStatus(d.ID, Error)
		return
	}
	incompleteFile.handle = handle
	defer handle.Close()
	if new {
		incompleteFile.prefill()
	}
	// TODO: Check if file length has changed if a .incomplete file was loaded
	// TODO: Check hash of file?
	for i, val := range incompleteFile.Splits {
		log.Printf("split index: %d, complete: %d\n", i, val.ChunksComplete)
		go incompleteFile.download(val)
	}
}

func (f *incompleteFile) prefill() {
	bufSize := f.ChunkSize
	b := make([]byte, bufSize)
	for i := f.TargetLength; i > 0; {
		if bufSize > i {
			bufSize = i
			b = make([]byte, bufSize)
		}
		f.handle.Write(b)
		i -= bufSize
	}
}

func (f *incompleteFile) download(s fileSplit) {
	// TODO
}
