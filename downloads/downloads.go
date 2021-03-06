package downloads

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"os"
	"path"
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
	ID           int            `json:"id"`
	Name         string         `json:"name"`
	Status       DownloadStatus `json:"status"`
	Downloaded   int            `json:"dl"`
	Total        int            `json:"total"`
	transferInfo transferInfo
}

type transferInfo struct {
	Url  string `json:"url"`
	Path string `json:"path"`
}

type DownloadStatus int

const (
	Error DownloadStatus = iota
	Queued
	Paused
	Prefilling
	Downloading
	Finished
	Cancelled
)

type downloadCache struct {
	Path string `json:"path"`
	Download
}

func New(opt Options, callback func(d Download)) (*Client, error) {
	downloads := make([]Download, 0)
	if file, err := os.ReadFile(path.Join(opt.Path, "downloads.json")); err == nil {
		data := make([]downloadCache, 0)
		if err = json.Unmarshal(file, &data); err == nil {
			for i, val := range data {
				val.Download.transferInfo = transferInfo{
					Path: val.Path,
				}
				downloads[i] = val.Download
			}
		} else {
			log.Println(err)
		}
	} else {
		log.Println(err)
	}
	return &Client{opt, callback, downloads, opt.MaxConcurrent, &sync.Mutex{}}, nil
}

func (c *Client) Shutdown() {
	for _, val := range c.WithStatus(Queued, Downloading, Prefilling) {
		c.Pause(val.ID)
	}
	// TODO: Save stuff to download cache file
}

func (c *Client) GetAll() []Download {
	return c.downloads
}

func (c *Client) Get(id int) (Download, error) {
	for _, val := range c.downloads {
		if val.ID == id {
			return val, nil
		}
	}
	return Download{}, errors.New("not found")
}

func (c *Client) Queue(url, path string) (id int) {
	c.mutex.Lock()
	id = rand.Intn(9000000)
	d := Download{
		ID:     id,
		Name:   filepath.Base(path),
		Status: Queued,
		transferInfo: transferInfo{
			Url:  url,
			Path: path,
		},
	}
	c.downloads = append(c.downloads, d)
	c.mutex.Unlock()
	c.callback(d)
	c.Run()
	return
}

func (c *Client) Retry(id int) {
	c.updateStatus(id, Queued)
	c.Run()
}

func (c *Client) Pause(id int) {
	c.updateStatus(id, Paused)
}

func (c *Client) Resume(id int) {
	c.updateStatus(id, Queued)
	if len(c.WithStatus(Downloading)) == 0 {
		c.Run()
	}
}

func (c *Client) Cancel(id int) {
	c.updateStatus(id, Cancelled)
}

func (c *Client) Run() {
	slots := c.slots
	if slots < 1 {
		return
	}
	// Find downloads to resume or start
	priority := c.WithStatus(Paused)
	queue := c.WithStatus(Queued)
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
		} else {
			val.Total = -1
			c.callback(val)
		}
	}
	c.downloads = d
	c.mutex.Unlock()
}
