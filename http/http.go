package http

import (
	"net/http"
	"github.com/sirupsen/logrus"
	"github.com/aspcartman/pcache"
)

type Handler struct {
	Cache *pcache.Cache
}

func (s *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	url := string(req.FormValue("url"))
	log := logrus.WithFields(logrus.Fields{
		"url": url,
	})
	log.Info("incoming request")

	if len(url) == 0 {
		rw.WriteHeader(http.StatusBadRequest)
		log.Error("empty url")
		return
	}

	img, cached, err := s.Cache.Get(url)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		log.WithError(err).Error("error acquiring image")
		return
	}

	log = log.WithField("cached", cached)
	_, err = rw.Write(img)
	if err != nil {
		log.WithError(err).Error("failed writing response")
		return
	}

	log.Info("done")
}
