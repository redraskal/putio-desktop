package downloads

import (
	"encoding/json"
	"os"
)

type cachedDownload struct {
	TransferInfo   transferInfo   `json:"transferInfo"`
	IncompleteFile incompleteFile `json:"incompleteFile"`
	Download
}

func readCache(path string) (cache []cachedDownload, err error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return
	}
	cache = make([]cachedDownload, 0)
	err = json.Unmarshal(file, &cache)
	return
}

func (c *Client) saveCache() error {
	data := make([]cachedDownload, len(c.downloads))
	for i, val := range c.downloads {
		data[i] = cachedDownload{
			val.transferInfo,
			*val.incompleteFile,
			*val,
		}
	}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(c.opt.cachePath, b, os.ModePerm)
}
