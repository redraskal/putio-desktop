package downloads

import (
	"log"
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
	log.Printf("Starting download for %d, %s\n", d.ID, d.url)
	c.updateStatus(d.ID, Downloading)
	c.mutex.Lock()
	c.slots--
	c.mutex.Unlock()
	header, err := Head(d.url)
	if err != nil {
		log.Println(err)
		c.updateStatus(d.ID, Error)
		return
	}
	log.Println("Len: ", header.ContentLength)
	// TODO: Load incompleteFile struct if it was paused
	// progress := incompleteFile{
	// 	ID: d.ID,
	// }
	// TODO: Create file and fill to total length
	for i := 0; i <= c.opt.Splits; i++ {
		// TODO
	}
}
