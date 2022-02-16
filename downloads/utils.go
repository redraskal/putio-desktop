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

func (c *Client) withStatus(s DownloadStatus) []Download {
	res := make([]Download, len(c.downloads))
	for _, val := range c.downloads {
		if val.Status == s {
			res = append(res, val)
		}
	}
	return res
}
