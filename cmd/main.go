package main

import (
	"flag"
	"github.com/aspcartman/pcache"
	"github.com/aspcartman/pcache/grace"
	phttp "github.com/aspcartman/pcache/http"
	"github.com/aspcartman/pcache/storage"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/aspcartman/pcache/e"
	"github.com/aspcartman/pcache/e/elogrus"
)

var log *logrus.Logger

type options struct {
	addr      string
	storePath string
}

func main() {
	initLogging()
	opts := flagParse()

	log.Info("starting")
	cls := start(opts.addr, opts.storePath)
	defer cls.Close()
	log.Info("started")

	waitForShutdown()
	log.Warning("shutting down")
}

func initLogging() {
	log = &logrus.Logger{
		Out: os.Stderr,
		Formatter: &logrus.TextFormatter{
			ForceColors:      true,
			DisableTimestamp: true,
		},
		Hooks: make(logrus.LevelHooks),
		Level: logrus.DebugLevel,
	}
	elogrus.AddLogger(log)
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

func start(addr, storePath string) *grace.Closer {
	cls := &grace.Closer{}
	defer e.OnError(func(e *e.Exception) {
		cls.Close()
	})

	log.WithField("path", storePath).Debug("opening store")
	store := openStore(storePath)
	defer cls.Add(func() {
		closeStore(store)
	})

	log.WithField("addr", addr).Debug("starting http server")
	lsn := startHTTP(addr, store)
	defer cls.Add(func() {
		stopHTTP(lsn)
	})

	return cls
}

/*
	Storage
 */
func openStore(path string) storage.Store {
	log := log.WithField("path", path)

	var store storage.Store
	var err error
	if len(path) == 0 {
		log.Debug("using inmemory store")
		store = storage.NewInmemStore()
	} else {
		log.Debug("opening badger store")
		store, err = storage.NewBadgerStore(path)
	}

	if err != nil {
		e.Throw("failed opening store", err)
	}

	return store
}

func closeStore(store storage.Store) {
	log.Warn("closing the store")
	if err := store.Close(); err != nil {
		log.WithError(err).Error("error during storage close")
		// ignore it
	}
	log.Warn("storage is closed")
}

/*
	HTTP
	Http is started with a graceful listener for the ability to
	gracefully shutdown
 */
func startHTTP(addr string, store storage.Store) *grace.GracefulListener {
	log := log.WithField("addr", addr)

	handler := &phttp.Handler{Cache: pcache.New(store), Log: log.Logger}
	srv := fasthttp.Server{
		Handler: handler.Serve,
	}

	rawListener, err := net.Listen("tcp", addr)
	if err != nil {
		log.WithError(err).Panic("failed listening")
	}

	gracefulListener := grace.NewGracefulListener(rawListener, 10*time.Second)

	go func() {
		// Inability to listen should crash the app
		log.Info("listening")
		if err := srv.Serve(gracefulListener); err != nil {
			e.Throw("failed listening&serving", err)
		}
		log.Warn("stopped listening")
	}()

	return gracefulListener
}

func stopHTTP(lsn *grace.GracefulListener) {
	log := log.WithField("addr", lsn.Addr())
	log.Warn("shutting down http server")

	if err := lsn.Close(); err != nil {
		log.WithError(err).Error("error during http shutdown")
		// ignore it
	}

	log.Warn("http server is down")
}
