package storage

import (
	. "github.com/glycerine/goconvey/convey"
	"os"
	"path/filepath"
	"testing"
)

func TestBoltStore(t *testing.T) {
	Convey("Having a simple store", t, func() {
		path := filepath.Join(os.TempDir(), "pcache/test/bolt")
		os.RemoveAll(path)
		os.MkdirAll(path, 0755)

		store, err := NewBoltStore(path+"/store.bolt")
		So(err, ShouldBeNil)

		testStore(store)
	})
}
