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
	imageData, err := c.getFromStore(url, size)
	if err == nil {
		return imageData, true
	} else if err != nil && err != storage.ErrNotFound {
		e.Throw(err)
	}

	// Maybe there is cached original?
	if size != SizeOrig {
		imageData, err = c.getFromStore(url, SizeOrig)
		if err != nil && err != storage.ErrNotFound {
			e.Throw(err)
		}
	}

	// Ok, let's download the image from source
	if len(imageData) == 0 {
		imageData = c.doGetData(url)
	}

	conv := c.convertImageToSize(imageData, size)
	c.saveToStore(url, size, imageData)

	go c.verifyProperlyCached(url, imageData)

	return conv, false
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

func (c *ImageCache) convertImageToSize(data []byte, size Size) []byte {
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
		if img.Bounds().Size().X > 256 || img.Bounds().Size().Y > 256 {
			img = resize.Thumbnail(256, 256, img, resize.Bicubic)
		}

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

func (c *ImageCache) verifyProperlyCached(url string, data []byte) {
	defer e.Catch(func(e *e.Exception) {})

	for _, sz := range []Size{SizeOrig, SizeSmall, SizePlaceholder} {
		_, err := c.getFromStore(url, sz)
		if err == storage.ErrNotFound {
			data = c.convertImageToSize(data, sz)
			c.saveToStore(url, sz, data)
		}
	}
}
