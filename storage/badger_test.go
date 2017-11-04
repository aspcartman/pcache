package storage

import (
	"testing"
	. "github.com/glycerine/goconvey/convey"
	"os"
	"path/filepath"
)

func TestBadgerStore(t *testing.T) {
	Convey("Having a simple store", t, func() {
		path := filepath.Join(os.TempDir(), "cache_proxy/test/badger")
		os.RemoveAll(path)
		os.MkdirAll(path, 0755)

		store, err := NewBadgerStore(path)
		So(err, ShouldBeNil)

		testStore(store)
	})
}
