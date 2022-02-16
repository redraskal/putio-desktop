package downloads

import (
	"log"
	"net/http"
	"strconv"
)

type incompleteFile struct {
	ID           int         `json:"id"`
	TargetLength int         `json:"len"`
	Splits       []fileSplit `json:"s"`
}

type fileSplit struct {
	ChunksComplete int `json:"c"`
}

func (c *Client) download(d Download) {
	if d.Status >= Downloading {
		return
	}
	c.updateStatus(d.ID, Downloading)
	c.mutex.Lock()
	c.slots--
	c.mutex.Unlock()
	// TODO: Download head to find file length
	fileLength, err := d.head()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Len: ", fileLength)
	// TODO: Load incompleteFile struct if it was paused
	// progress := incompleteFile{
	// 	ID: d.ID,

	// }
	// TODO: Create file and fill to total length
	for i := 0; i <= c.opt.Splits; i++ {
		// TODO
	}
}

func (d Download) head() (fileLength int, err error) {
	res, err := http.Head(d.URL)
	if err != nil {
		return
	}
	val := res.Header.Get("Content-Length")
	return strconv.Atoi(val)
}
