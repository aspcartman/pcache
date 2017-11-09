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
	defer panicHandler(ctx)

	url, size := s.getArgs(ctx)
	log := s.Log.WithFields(logrus.Fields{
		"url":  url,
		"size": size,
	})

	s.verifyArgs(url, size)

	img, cached, err := s.Cache.Get(url, size)
	if err != nil {
		e.Throw("error acquiring image", log, err)
	}

	log = log.WithField("cached", cached)
	_, err = ctx.Write(img)
	if err != nil {
		e.Throw("failed writing response", log, err)
	}

	log.Info("done")
}

func panicHandler(ctx *fasthttp.RequestCtx) {
	e.Catch(func(ex *e.Exception) {
		switch ex.Error {
		case ErrBadRequest:
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
		default:
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		}
		ctx.WriteString(ex.Info())
	})
}

func (s *Handler) getArgs(ctx *fasthttp.RequestCtx) (url string, size pcache.Size) {
	return string(ctx.QueryArgs().Peek("url")), pcache.Size(ctx.QueryArgs().Peek("size"))
}

func (s *Handler) verifyArgs(url string, size pcache.Size) {
	if len(url) == 0 {
		e.Throw(ErrBadRequest, "bad url")
	}

	switch size {
	case pcache.SizeOrig:
	case pcache.SizeSmall:
	default:
		e.Throw(ErrBadRequest, "bad size")
	}
}
