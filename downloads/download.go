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
	ID           int
	TargetLength int
	ChunkSize    int
	Splits       []*fileSplit
	handle       *os.File
	transferInfo transferInfo
}

type fileSplit struct {
	ChunksComplete int
	Start          int
	End            int
	NChunks        int
}

const MIN_LEN_TO_SPLIT = 100 * 1e+6 // 100 mb

func (c *Client) download(d *Download) {
	if d.Status >= Prefilling {
		return
	}
	new := d.Downloaded == 0
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
	if !new && d.incompleteFile.TargetLength != header.ContentLength {
		log.Println(errors.New("head len does not match target len"))
		c.updateStatus(d.ID, Error)
		return
	}
	if !new {
		if stat, err := os.Stat(d.Path()); err != nil || stat.Size() != int64(d.incompleteFile.TargetLength) {
			log.Printf("Restarting download because file len does not match target len for %d, %s\n", d.ID, d.transferInfo.Url)
			c.updateDownloaded(d.ID, 0)
			d.incompleteFile.transferInfo = d.transferInfo
			new = true
		}
	}
	nSplits := c.opt.Splits
	if header.ContentLength <= c.opt.Splits || header.ContentLength < MIN_LEN_TO_SPLIT {
		nSplits = 1
	}
	if new {
		d.incompleteFile.TargetLength = header.ContentLength
		d.incompleteFile.ChunkSize = 30 * 1e+6 // 30 mb
		var s []*fileSplit
		for i := 0; i < nSplits; i++ {
			s = append(s, &fileSplit{})
		}
		d.incompleteFile.Splits = s
	} else {
		d.incompleteFile.transferInfo = d.transferInfo
	}
	handle, err := os.OpenFile(d.Path(), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Println(err)
		c.updateStatus(d.ID, Error)
		return
	}
	d.incompleteFile.handle = handle
	if new {
		c.updateStatus(d.ID, Prefilling)
		d.incompleteFile.prefill(c)
		c.updateStatus(d.ID, Downloading)
		c.updateDownloaded(d.ID, 0)
	}
	defer func() {
		c.mutex.Lock()
		c.slots++
		c.mutex.Unlock()
		c.Run()
	}()
	splitSize := d.incompleteFile.TargetLength / nSplits
	var g errgroup.Group
	for i, val := range d.incompleteFile.Splits {
		if new {
			val.Start = i * splitSize
			val.End = val.Start + splitSize
			if i == (nSplits - 1) {
				val.End = d.incompleteFile.TargetLength
			}
			val.NChunks = (val.End - val.Start) / d.incompleteFile.ChunkSize
			if val.NChunks < 1 {
				val.NChunks = 1
			}
		}
		// Turns i & val into constants so they are not modified due to delayed download calls.
		iC := i
		valC := val
		g.Go(func() error {
			return d.incompleteFile.download(valC, iC, d.ID, c)
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
	if d.Status == Cancelled {
		log.Printf("Download cancelled for %d, %s", d.ID, d.transferInfo.Url)
		c.clear(Cancelled)
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
	log.Printf("split index: %d, start: %d, end: %d, complete: %d/%d\n", splitIndex, s.Start, s.End, s.ChunksComplete, s.NChunks)
	for i := s.ChunksComplete; i < s.NChunks; i++ {
		if d, err := c.Get(id); err != nil || (d.Status < Prefilling || d.Status > Downloading) {
			log.Printf("split index: %d, paused\n", splitIndex)
			break
		}
		start := s.Start + (i * f.ChunkSize)
		end := start + f.ChunkSize
		if i == (s.NChunks - 1) {
			end = s.End
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
		if res.StatusCode != http.StatusOK && (s.NChunks > 1 && res.StatusCode != http.StatusPartialContent) {
			return errors.New("split: bad status: " + res.Status)
		}
		w := NewSectionWriter(f.handle, int64(start), int64(size+1))
		if _, err := io.Copy(w, res.Body); err != nil {
			return err
		}
		s.ChunksComplete++
		c.incrementDownloaded(id, size)
		log.Printf("split index: %d, complete: %d/%d\n", splitIndex, s.ChunksComplete, s.NChunks)
	}
	return nil
}
