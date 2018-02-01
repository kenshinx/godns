package main

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_Suffix_Tree(t *testing.T) {
	root := newSuffixTreeRoot()

	Convey("Google should not be found", t, func() {
		root.insert("cn", "114.114.114.114")
		root.sinsert([]string{"baidu", "cn"}, "166.111.8.28")
		root.sinsert([]string{"sina", "cn"}, "114.114.114.114")

		v, found := root.search(strings.Split("google.com", "."))
		So(found, ShouldEqual, false)

		v, found = root.search(strings.Split("baidu.cn", "."))
		So(found, ShouldEqual, true)
		So(v, ShouldEqual, "166.111.8.28")
	})

	Convey("Google should be found", t, func() {
		root.sinsert(strings.Split("com", "."), "")
		root.sinsert(strings.Split("google.com", "."), "8.8.8.8")
		root.sinsert(strings.Split("twitter.com", "."), "8.8.8.8")
		root.sinsert(strings.Split("scholar.google.com", "."), "208.67.222.222")

		v, found := root.search(strings.Split("google.com", "."))
		So(found, ShouldEqual, true)
		So(v, ShouldEqual, "8.8.8.8")

		v, found = root.search(strings.Split("www.google.com", "."))
		So(found, ShouldEqual, true)
		So(v, ShouldEqual, "8.8.8.8")

		v, found = root.search(strings.Split("scholar.google.com", "."))
		So(found, ShouldEqual, true)
		So(v, ShouldEqual, "208.67.222.222")

		v, found = root.search(strings.Split("twitter.com", "."))
		So(found, ShouldEqual, true)
		So(v, ShouldEqual, "8.8.8.8")

		v, found = root.search(strings.Split("baidu.cn", "."))
		So(found, ShouldEqual, true)
		So(v, ShouldEqual, "166.111.8.28")
	})

}
