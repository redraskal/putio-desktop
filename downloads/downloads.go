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
	downloads []*Download
	slots     int
	mutex     *sync.Mutex
}

type Options struct {
	Path          string
	MaxConcurrent int
	Splits        int
}

type Download struct {
	ID             int            `json:"id"`
	Name           string         `json:"name"`
	Status         DownloadStatus `json:"status"`
	Downloaded     int            `json:"dl"`
	Total          int            `json:"total"`
	transferInfo   transferInfo
	incompleteFile *incompleteFile
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
	TransferInfo   transferInfo   `json:"t"`
	IncompleteFile incompleteFile `json:"if"`
	Download
}

func New(opt Options, callback func(d Download)) (*Client, error) {
	downloads := make([]*Download, 0)
	if file, err := os.ReadFile(path.Join(opt.Path, "downloads.json")); err == nil {
		data := make([]downloadCache, 0)
		if err = json.Unmarshal(file, &data); err == nil {
			for _, val := range data {
				val.Download.transferInfo = val.TransferInfo
				val.Download.incompleteFile = &val.IncompleteFile
				downloads = append(downloads, &val.Download)
			}
		} else {
			log.Println(err)
		}
	} else {
		log.Println(err)
	}
	return &Client{opt, callback, downloads, opt.MaxConcurrent, &sync.Mutex{}}, nil
}

func (c *Client) Shutdown() error {
	for _, val := range c.WithStatus(Queued, Downloading, Prefilling) {
		c.Pause(val.ID)
	}
	// Save downloads to download cache file
	data := make([]downloadCache, len(c.downloads))
	for i, val := range c.downloads {
		data[i] = downloadCache{
			val.transferInfo,
			*val.incompleteFile,
			*val,
		}
	}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path.Join(c.opt.Path, "downloads.json"), b, os.ModePerm)
}

func (c *Client) GetAll() []Download {
	d := make([]Download, 0)
	for _, val := range c.downloads {
		if val.ID != 0 {
			d = append(d, *val)
		}
	}
	return d
}

func (c *Client) Get(id int) (Download, error) {
	for _, val := range c.downloads {
		if val.ID == id {
			return *val, nil
		}
	}
	return Download{}, errors.New("not found")
}

func (c *Client) Queue(url, path string) (id int) {
	c.mutex.Lock()
	id = rand.Intn(9000000)
	t := transferInfo{
		Url:  url,
		Path: path,
	}
	d := Download{
		ID:           id,
		Name:         filepath.Base(path),
		Status:       Queued,
		transferInfo: t,
		incompleteFile: &incompleteFile{
			ID:           id,
			transferInfo: t,
		},
	}
	c.downloads = append(c.downloads, &d)
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

func (c *Client) Cancel(id int) error {
	d, err := c.Get(id)
	if err != nil {
		return err
	}
	if d.Status < Prefilling {
		c.updateStatus(id, Cancelled)
		c.clear(Cancelled)
	} else {
		c.updateStatus(id, Cancelled)
	}
	return nil
}

func (c *Client) Run() {
	slots := c.slots
	if slots < 1 {
		return
	}
	queue := c.WithStatus(Queued)
	for _, val := range queue {
		if slots < 1 {
			break
		}
		go c.download(val)
		slots--
	}
}

func (c *Client) RunAndResume() {
	slots := c.slots
	if slots < 1 {
		return
	}
	priority := c.WithStatus(Paused)
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
	c.Run()
}

func (c *Client) clear(s DownloadStatus) {
	c.mutex.Lock()
	d := c.downloads[:0]
	for _, val := range c.downloads {
		if val.Status != s {
			d = append(d, val)
		} else {
			val.Total = -1
			c.callback(*val)
		}
	}
	c.downloads = d
	c.mutex.Unlock()
}

func (c *Client) ClearCompleted() {
	c.clear(Finished)
}
