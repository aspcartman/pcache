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
)

var log *logrus.Logger

type options struct {
	addr      string
	storePath string
}

func main() {
	initLogging()

	log.Info("starting")
	cls := start(flagParse())
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

func start(opts options) *grace.Closer {
	cls := &grace.Closer{}
	defer func() {
		if r := recover(); r != nil {
			cls.Close()
		}
	}()

	log.WithField("path", opts.storePath).Debug("opening store")
	store := openStore(opts.storePath)
	defer cls.Add(func() {
		closeStore(store)
	})

	log.WithField("addr", opts.addr).Debug("starting http server")
	lsn := startHTTP(opts.addr, store)
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
		log.Fatal("failed opening store")
	}

	return store
}

func closeStore(store storage.Store) {
	log.Warn("closing the store")
	if err := store.Close(); err != nil {
		log.WithError(err).Error("error during storage close")
	}
	log.Warn("storage is closed")
}

/*
	HTTP
 */
func startHTTP(addr string, store storage.Store) *grace.GracefulListener {
	log := log.WithField("addr", addr)

	handler := &phttp.Handler{Cache: pcache.New(store), Log:log}
	srv := fasthttp.Server{
		Handler: handler.Serve,
	}

	rawListener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("failed listening")
	}

	gracefulListener := grace.NewGracefulListener(rawListener, 10*time.Second)

	go func() {
		log.Info("listening")
		if err := srv.Serve(gracefulListener); err != nil {
			log.WithError(err).Panic("failed listening & serving")
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
	}

	log.Warn("http server is down")
}
