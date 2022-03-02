package downloads

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"golang.org/x/sync/errgroup"
)

type incompleteFile struct {
	ID           int         `json:"id"`
	TargetLength int         `json:"len"`
	ChunkSize    int         `json:"c"`
	Splits       []fileSplit `json:"s"`
	handle       *os.File
	transferInfo transferInfo
}

type fileSplit struct {
	ChunksComplete int `json:"c"`
	start          int
	end            int
	nChunks        int
}

const MIN_LEN_TO_SPLIT = 100 * 1e+6 // 100 mb

func (c *Client) download(d Download) {
	if d.Status >= Prefilling {
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
	c.updateTotal(d.ID, header.ContentLength)
	nSplits := c.opt.Splits
	if header.ContentLength <= c.opt.Splits || header.ContentLength < MIN_LEN_TO_SPLIT {
		nSplits = 1
	}
	incompleteFile := incompleteFile{
		ID:           d.ID,
		TargetLength: header.ContentLength,
		ChunkSize:    30 * 1e+6, // 30 mb
		Splits:       make([]fileSplit, nSplits),
		transferInfo: d.transferInfo,
	}
	handle, err := os.OpenFile(d.Path(), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Println(err)
		c.updateStatus(d.ID, Error)
		return
	}
	incompleteFile.handle = handle
	if new {
		c.updateStatus(d.ID, Prefilling)
		incompleteFile.prefill(c)
		c.updateStatus(d.ID, Downloading)
		c.updateDownloaded(d.ID, 0)
	}
	defer func() {
		c.mutex.Lock()
		c.slots++
		c.mutex.Unlock()
		c.Run()
	}()
	// TODO: Check if file length has changed if a .incomplete file was loaded
	splitSize := incompleteFile.TargetLength / nSplits
	var g errgroup.Group
	for i, val := range incompleteFile.Splits {
		val.start = i * splitSize
		val.end = val.start + splitSize
		if i == (nSplits - 1) {
			val.end = incompleteFile.TargetLength
		}
		val.nChunks = (val.end - val.start) / incompleteFile.ChunkSize
		if val.nChunks < 1 {
			val.nChunks = 1
		}
		// Turns i & val into constants so they are not modified due to delayed download calls.
		iC := i
		valC := val
		g.Go(func() error {
			return incompleteFile.download(&valC, iC, d.ID, c)
		})
	}
	if err := g.Wait(); err != nil {
		handle.Close()
		log.Printf("Download failed for %d, %s: %s\n", d.ID, d.transferInfo.Url, err)
		c.updateStatus(d.ID, Error)
		return
	}
	handle.Close()
	if d.Status == Paused {
		log.Printf("Download paused for %d, %s", d.ID, d.transferInfo.Url)
		return
	}
	log.Printf("Download complete for %d, %s\n", d.ID, d.transferInfo.Url)
	c.updateStatus(d.ID, Finished)
	if err := os.Rename(d.transferInfo.Path+".incomplete", d.transferInfo.Path); err != nil {
		log.Println(err)
	}
}

func (f *incompleteFile) prefill(c *Client) {
	bufSize := f.ChunkSize
	b := make([]byte, bufSize)
	for i := f.TargetLength; i > 0; {
		if bufSize > i {
			bufSize = i
			b = make([]byte, bufSize)
		}
		f.handle.Write(b)
		c.incrementDownloaded(f.ID, bufSize)
		i -= bufSize
	}
}

func (f *incompleteFile) download(s *fileSplit, splitIndex int, id int, c *Client) error {
	log.Printf("split index: %d, start: %d, end: %d, complete: %d/%d\n", splitIndex, s.start, s.end, s.ChunksComplete, s.nChunks)
	for i := s.ChunksComplete; i < s.nChunks; i++ {
		if d, err := c.Get(id); err != nil || (d.Status < Prefilling || d.Status > Downloading) {
			log.Printf("split index: %d, paused\n", splitIndex)
			break
		}
		start := s.start + (i * f.ChunkSize)
		end := start + f.ChunkSize
		if i == (s.nChunks - 1) {
			end = s.end
		}
		size := (end - start)
		r := fmt.Sprintf("bytes=%d-%d", start, end)
		req, err := http.NewRequest("GET", f.transferInfo.Url, nil)
		if err != nil {
			return err
		}
		req.Header.Add("Range", r)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK && (s.nChunks > 1 && res.StatusCode != http.StatusPartialContent) {
			return errors.New("split: bad status: " + res.Status)
		}
		w := NewSectionWriter(f.handle, int64(start), int64(size+1))
		if _, err := io.Copy(w, res.Body); err != nil {
			return err
		}
		s.ChunksComplete++
		c.incrementDownloaded(id, size)
		log.Printf("split index: %d, complete: %d/%d\n", splitIndex, s.ChunksComplete, s.nChunks)
	}
	return nil
}
