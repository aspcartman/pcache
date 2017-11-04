package storage

import (
	"testing"
	. "github.com/glycerine/goconvey/convey"
)

func TestInMemStore(t *testing.T) {
	Convey("Having inmemory store", t, func() {
		testStore(NewInmemStore())
	})
}
