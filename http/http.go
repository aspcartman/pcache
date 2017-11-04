package http

import (
	"github.com/aspcartman/pcache"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"net/http"
)

// Handler serves HTTP requests with "url" query arg
type Handler struct {
	Cache *pcache.Cache
}

func (s *Handler) Serve(ctx *fasthttp.RequestCtx) {
	url := string(ctx.QueryArgs().Peek("url"))
	log := logrus.WithFields(logrus.Fields{
		"url": url,
	})
	log.Info("incoming request")

	if len(url) == 0 {
		ctx.SetStatusCode(http.StatusBadRequest)
		log.Error("empty url")
		return
	}

	img, cached, err := s.Cache.Get(url)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		log.WithError(err).Error("error acquiring image")
		return
	}

	log = log.WithField("cached", cached)
	_, err = ctx.Write(img)
	if err != nil {
		log.WithError(err).Error("failed writing response")
		return
	}

	log.Info("done")
}
