package pcache

import (
	"fmt"
	"github.com/aspcartman/pcache/storage"
	"net/http"
	"time"
	"github.com/valyala/fasthttp"
	"github.com/nfnt/resize"
	"image"
	"bytes"
	"image/jpeg"
)

type Size string

const (
	SizeOrig  Size = "orig"
	SizeSmall Size = "small"
)

type ImageCache struct {
	client fasthttp.Client
	store  storage.Store
}

func New(store storage.Store) *ImageCache {
	return &ImageCache{
		client: fasthttp.Client{},
		store:  store,
	}
}

func (c *ImageCache) Get(url string, size Size) ([]byte, bool, error) {
	// Try to get image from cache
	img, err := c.getFromStore(url, size)
	if err == nil {
		return img, true, nil
	} else if err != nil && err != storage.ErrNotFound {
		return nil, false, err
	}

	// Maybe there is cached original?
	if size != SizeOrig {
		img, err = c.getFromStore(url, SizeOrig)
		if err != nil && err != storage.ErrNotFound {
			return nil, false, err
		}
	}

	// Ok, let's download the image from source
	if len(img) == 0 {
		img, err = c.doGetData(url)
		if err != nil {
			return nil, false, err
		}
		// Don't forget to save the original
		c.saveToStore(url, SizeOrig, img)
	}

	img, err = c.resizeImage(img, size)
	if err != nil {
		return nil, false, err
	}

	err = c.saveToStore(url, size, img)
	if err != nil {
		return nil, false, err
	}

	return img, false, nil
}

func (c *ImageCache) getFromStore(url string, size Size) ([]byte, error) {
	return c.store.Get(url + "_" + string(size))
}

func (c *ImageCache) saveToStore(url string, size Size, data []byte) error {
	return c.store.Set(url+"_"+string(size), data)
}

func (c *ImageCache) doGetData(url string) ([]byte, error) {
	code, res, err := c.client.GetTimeout(nil, url, 10*time.Second)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", code)
	}

	return res, nil
}

// warn: modifies the passed slice's data
func (c *ImageCache) resizeImage(data []byte, size Size) ([]byte, error) {
	if size == SizeOrig {
		return data, nil
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	img = resize.Thumbnail(200, 200, img, resize.Lanczos3)
	buf := bytes.NewBuffer(data[:0])
	err = jpeg.Encode(buf, img, &jpeg.Options{80})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
