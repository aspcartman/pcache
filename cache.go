package pcache

import (
	"github.com/aspcartman/pcache/storage"
	"net/http"
	"time"
	"github.com/valyala/fasthttp"
	"github.com/nfnt/resize"
	"image"
	"bytes"
	"image/jpeg"
	"github.com/fogleman/primitive/primitive"
	"runtime"
	_ "image/png"
	"github.com/aspcartman/pcache/e"
	"strings"
)

type Size string

const (
	SizeOrig        Size = "orig"
	SizeSmall       Size = "small"
	SizePlaceholder Size = "placeholder"
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

func (c *ImageCache) Get(url string, size Size) ([]byte, bool) {
	// Try to get image from cache
	img, err := c.getFromStore(url, size)
	if err == nil {
		return img, true
	} else if err != nil && err != storage.ErrNotFound {
		e.Throw(err)
	}

	// Maybe there is cached original?
	if size != SizeOrig {
		img, err = c.getFromStore(url, SizeOrig)
		if err != nil && err != storage.ErrNotFound {
			e.Throw(err)
		}
	}

	// Ok, let's download the image from source
	if len(img) == 0 {
		img = c.doGetData(url)
	}

	var requested []byte
	for _, sz := range []Size{SizeOrig, SizeSmall, SizePlaceholder} {
		img = c.resizeImage(img, sz)
		c.saveToStore(url, sz, img)
		if sz == size {
			requested = img
		}
	}

	if len(requested) == 0 {
		e.Throw("final image is empty")
	}

	return requested, false
}

func (c *ImageCache) Cache(url string) {
	img, err := c.getFromStore(url, SizeOrig)
	if err != nil && err != storage.ErrNotFound {
		e.Throw(err)
	}

	if len(img) == 0 {
		img = c.doGetData(url)
	}

	for _, sz := range []Size{SizeOrig, SizeSmall, SizePlaceholder} {
		_, err := c.getFromStore(url, sz)
		if err == nil {
			continue
		}
		if err != nil && err != storage.ErrNotFound {
			e.Throw(err)
		}

		img = c.resizeImage(img, sz)
		c.saveToStore(url, sz, img)
	}
}

func (c *ImageCache) getFromStore(url string, size Size) ([]byte, error) {
	return c.store.Get(url + "_" + string(size))
}

func (c *ImageCache) saveToStore(url string, size Size, data []byte) {
	e.Must(c.store.Set(url+"_"+string(size), data))
}

func (c *ImageCache) doGetData(url string) []byte {
	code, res, err := c.client.GetTimeout(nil, url, 10*time.Second)
	if err != nil {
		e.Throw(err, "failed getting img from source", url)
	}
	if code != http.StatusOK {
		e.Throw("bad status code", code)
	}

	return res
}

// warn: modifies the passed slice's data
func (c *ImageCache) resizeImage(data []byte, size Size) []byte {
	if size == SizeOrig {
		return data
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		e.Throw("failed decoding image", err, string(data))
	}

	switch size {

	case SizeSmall:
		img = resize.Thumbnail(256, 256, img, resize.Bicubic)
		buf := bytes.NewBuffer(nil)
		err = jpeg.Encode(buf, img, &jpeg.Options{80})
		if err != nil {
			e.Throw("failed encoding", err)
		}
		return buf.Bytes()

	case SizePlaceholder:
		bg := primitive.MakeColor(primitive.AverageImageColor(img))
		model := primitive.NewModel(img, bg, 256, runtime.NumCPU())
		for i := 0; i < 100; i++ {
			model.Step(primitive.ShapeTypeTriangle, 128, 0)
		}
		svg := model.SVG()
		svg = strings.Replace(svg, "<g", `<filter id="b"><feGaussianBlur stdDeviation="12" /></filter><g filter="url(#b)"`, 1)
		return []byte(svg)
	}

	e.Throw("unknown imgsize", size)
	return nil
}
