package downloads

import "net/http"

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
	// TODO
	return
}
