package elogrus

import (
	"github.com/sirupsen/logrus"
	"github.com/aspcartman/pcache/e"
)

// Adds a logrus hook, that prints all thrown exceptions
// to the logger provided. It allows you to pass logrus.Entry
// to the Throw() call.
func AddLogger(l *logrus.Logger) {
	e.RegisterPostHook(func(ex *e.Exception) {
		log := l.WithError(ex.Error)
		if entry := extractLogEntry(ex); entry != nil {
			log = log.WithFields(entry.Data)
		}
		log.Error(ex.Info())
	})
}

func extractLogEntry(exception *e.Exception) *logrus.Entry {
	desc := exception.Description
	for i, d := range desc {
		if entry, ok := d.(*logrus.Entry); ok {
			exception.Description = desc[:i+copy(desc[i:], desc[i+1:])] // delete the log entry from exception description
			return entry
		}
	}
	return nil
}
