package storage

import (
	. "github.com/glycerine/goconvey/convey"
	"testing"
)

func TestInMemStore(t *testing.T) {
	Convey("Having inmemory store", t, func() {
		testStore(NewInmemStore())
	})
}
