package downloads

func (c *Client) updateStatus(id int, s DownloadStatus) {
	for i, val := range c.downloads {
		if val.ID == id {
			c.mutex.Lock()
			c.downloads[i].Status = s
			c.mutex.Unlock()
			c.callback(c.downloads[i])
			break
		}
	}
}

func (c *Client) WithStatus(s DownloadStatus) []Download {
	res := make([]Download, 0)
	for _, val := range c.downloads {
		if val.Status == s {
			res = append(res, val)
		}
	}
	return res
}
