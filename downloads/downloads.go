package downloads

import (
	"io"
	"math/rand"
	"path/filepath"
	"sync"
)

type Client struct {
	opt       Options
	callback  func(d Download)
	downloads []Download
	slots     int
	mutex     *sync.Mutex
}

type Options struct {
	Path          string
	MaxConcurrent int
	Splits        int
}

type Download struct {
	ID     int            `json:"id"`
	Name   string         `json:"name"`
	Status DownloadStatus `json:"status"`
	url    string
	path   string
	writer *io.Writer
}

type DownloadStatus int

const (
	Error DownloadStatus = iota
	Queued
	Paused
	Downloading
	Finished
	Cancelled
)

func New(opt Options, callback func(d Download)) (*Client, error) {
	// TODO: Load download states from file
	// TODO: Add shutdown hook
	return &Client{opt, callback, []Download{}, opt.MaxConcurrent, &sync.Mutex{}}, nil
}

func (c *Client) Get() []Download {
	return c.downloads
}

func (c *Client) Queue(url, path string) (id int) {
	c.mutex.Lock()
	id = rand.Int()
	d := Download{
		ID:     id,
		Name:   filepath.Base(path),
		Status: Queued,
		url:    url,
		path:   path,
	}
	c.downloads = append(c.downloads, d)
	c.mutex.Unlock()
	c.callback(d)
	if len(c.withStatus(Downloading)) == 0 {
		c.Run()
	}
	return
}

func (c *Client) Retry(id int) {
	c.updateStatus(id, Queued)
	if len(c.withStatus(Downloading)) == 0 {
		c.Run()
	}
}

func (c *Client) Pause(id int) {
	c.updateStatus(id, Paused)
}

func (c *Client) Resume(id int) {
	c.updateStatus(id, Queued)
	if len(c.withStatus(Downloading)) == 0 {
		c.Run()
	}
}

func (c *Client) Cancel(id int) {
	c.updateStatus(id, Cancelled)
}

func (c *Client) Run() {
	// Find downloads to resume or start
	priority := c.withStatus(Paused)
	queue := c.withStatus(Queued)
	slots := c.slots
	for _, val := range priority {
		if slots < 1 {
			break
		}
		go c.download(val)
		slots--
	}
	if slots < 1 {
		return
	}
	for _, val := range queue {
		if slots < 1 {
			break
		}
		go c.download(val)
		slots--
	}
}

func (c *Client) ClearCompleted() {
	c.mutex.Lock()
	d := c.downloads[:0]
	for _, val := range c.downloads {
		if val.Status != Finished {
			d = append(d, val)
		}
	}
	c.downloads = d
	c.mutex.Unlock()
}
