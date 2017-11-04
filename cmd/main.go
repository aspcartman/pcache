package main

import (
	"context"
	"flag"
	phttp "github.com/aspcartman/pcache/http"
	"github.com/aspcartman/pcache/storage"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/aspcartman/pcache"
)

type options struct {
	addr      string
	storePath string
}

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:true,
		DisableTimestamp:true,
	})
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Info("starting")

	service, err := newService(flagParse())
	if err != nil {
		logrus.WithError(err).Fatal("failed stating")
	}

	go service.Serve()

	waitForShutdown()
	logrus.Warning("shutting down")
	service.Stop()
	logrus.Warning("down.")
}

func flagParse() options {
	opts := options{}
	flag.StringVar(&opts.addr, "addr", "0.0.0.0:8080", "addr to listen on")
	flag.StringVar(&opts.storePath, "store", "", "path to the store, defaults to inmemory store")
	flag.Parse()
	return opts
}

func waitForShutdown() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}

type service struct {
	store      storage.Store
	httpServer *http.Server
}

func newService(opts options) (*service, error) {
	store, err := openStore(opts.storePath)
	if err != nil {
		return nil, err
	}

	httpSrv := &http.Server{
		Handler: &phttp.Handler{Cache:pcache.New(store)},
		Addr:    opts.addr,
	}

	return &service{store, httpSrv}, nil
}

func openStore(path string) (storage.Store, error) {
	if len(path) == 0 {
		logrus.Debug("using inmemory store")
		return storage.NewInmemStore(), nil
	} else {
		logrus.WithField("path", path).Debug("opening badger store")
		return storage.NewBadgerStore(path)
	}
}

func (s *service) Serve() {
	log := logrus.WithField("addr", s.httpServer.Addr)
	log.Info("listening")
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.WithError(err).Panic("failed listening & serving")
	}
	log.Warn("stopped listening")
}

func (s *service) Stop() {
	s.shutdownHTTP()
	s.shutdownStore()
}

func (s *service) shutdownHTTP() {
	log := logrus.WithField("addr", s.httpServer.Addr)
	log.Warn("shutting down http server")

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.WithError(err).Error("error during http shutdown")
	}

	log.Warn("http server is down")
}

func (s *service) shutdownStore() {
	logrus.Warn("closing the store")
	if err := s.store.Close(); err != nil {
		logrus.WithError(err).Error("error during storage close")
	}
	logrus.Warn("storage is closed")
}
