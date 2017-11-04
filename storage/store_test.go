package storage

import (
	. "github.com/glycerine/goconvey/convey"
)

func testStore(store Store) {
	Convey("Get non existing key fails with appropriate error", func() {
		_, err := store.Get("lol")
		So(err, ShouldEqual, ErrNotFound)
	})

	Convey("Setting works", func() {
		err := store.Set("lol", []byte{1, 2, 3})
		So(err, ShouldBeNil)

		Convey("And getting it all back too", func() {
			v, err := store.Get("lol")
			So(err, ShouldBeNil)
			So(v, ShouldResemble, []byte{1, 2, 3})
		})
	})
}