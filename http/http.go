package http

import (
	"github.com/aspcartman/pcache"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"github.com/aspcartman/pcache/e"
	"github.com/pkg/errors"
)

/*
	Basic HTTP handler for image caching.
	Can be used as is or as an example.

	QueryArgs:
		url - absolute url for the original image
		size - size of the requested image (pcache.Size enum)

	Writes requested image bytes in response
 */

type Handler struct {
	Cache *pcache.ImageCache
	Log   *logrus.Logger
}

var ErrBadRequest = errors.New("bad request")

func (s *Handler) Serve(ctx *fasthttp.RequestCtx) {
	defer e.Catch(func(ex *e.Exception) {
		switch ex.Error {
		case ErrBadRequest:
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
		default:
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		}
		ctx.WriteString(ex.Info())
	})

	url, size, onlyCache := s.getArgs(ctx)
	log := s.Log.WithFields(logrus.Fields{
		"url":       url,
		"size":      size,
		"onlyCache": onlyCache,
	})

	s.verifyArgs(url, size)

	if onlyCache {
		go func() {
			e.Catch(func(e *e.Exception) {})
			s.Cache.Cache(url)
		}()
		ctx.WriteString("Accepted")
	} else {
		img, cached := s.Cache.Get(url, size)
		log = log.WithField("cached", cached)
		_, err := ctx.Write(img)
		if err != nil {
			e.Throw("failed writing response", log, err)
		}
	}

	log.Info("done")
}

func (s *Handler) getArgs(ctx *fasthttp.RequestCtx) (url string, size pcache.Size, onlyCache bool) {
	return string(ctx.QueryArgs().Peek("url")), pcache.Size(ctx.QueryArgs().Peek("size")), ctx.QueryArgs().Has("onlyCache")
}

func (s *Handler) verifyArgs(url string, size pcache.Size) {
	if len(url) == 0 {
		e.Throw(ErrBadRequest, "bad url")
	}

	switch size {
	case pcache.SizeOrig:
	case pcache.SizeSmall:
	case pcache.SizePlaceholder:
	default:
		e.Throw(ErrBadRequest, "bad size")
	}
}
