package pcache

import (
	"fmt"
	"github.com/aspcartman/pcache/storage"
	"net/http"
	"time"
	"github.com/valyala/fasthttp"
)

type Cache struct {
	client fasthttp.Client
	store  storage.Store
}

func New(store storage.Store) *Cache {
	return &Cache{
		client: fasthttp.Client{},
		store:  store,
	}
}

func (c *Cache) Get(url string) ([]byte, bool, error) {
	if img, err := c.store.Get(url); err == nil {
		return img, true, nil
	} else if err != nil && err != storage.ErrNotFound {
		return nil, false, err
	}

	if img, err := c.doGetData(url); err == nil {
		return img, false, c.store.Set(url, img)
	} else {
		return nil, false, err
	}
}

func (c *Cache) doGetData(url string) ([]byte, error) {
	code, res, err := c.client.GetTimeout(nil, url, 10*time.Second)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", code)
	}

	return res, nil
}
