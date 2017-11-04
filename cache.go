package pcache

import (
	"fmt"
	"github.com/aspcartman/pcache/storage"
	"time"
	"net/http"
	"io"
	"io/ioutil"
)

type Cache struct {
	client http.Client
	store  storage.Store
}

func New(store storage.Store) *Cache {
	return &Cache{
		client: http.Client{Timeout: 10 * time.Second},
		store:  store,
	}
}

func (c *Cache) Get(url string) ([]byte, bool, error) {
	if img, err := c.store.Get(url); err == nil {
		return img, true, nil
	} else if err != nil && err != storage.ErrNotFound {
		return nil, false, err
	}

	if img, err := doGetData(url); err == nil {
		return img, false, c.store.Set(url, img)
	} else {
		return nil, false, err
	}
}

func doGetData(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", res.StatusCode)
	}

	defer func() {
		ioutil.ReadAll(res.Body)
		res.Body.Close()
	}()

	data := make([]byte, res.ContentLength)
	_, err = io.ReadFull(res.Body, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
