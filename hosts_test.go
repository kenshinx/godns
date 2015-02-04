package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHostDomainAndIP(t *testing.T) {
	Convey("Test Host File Domain and IP regex", t, func() {
		f := &FileHosts{}

		Convey("1.1.1.1 should be IP and not domain", func() {
			So(f.isIP("1.1.1.1"), ShouldEqual, true)
			So(f.isDomain("1.1.1.1"), ShouldEqual, false)
		})

		Convey("2001:470:20::2 should be IP and not domain", func() {
			So(f.isIP("2001:470:20::2"), ShouldEqual, true)
			So(f.isDomain("2001:470:20::2"), ShouldEqual, false)
		})

		Convey("`host` should be domain and not IP", func() {
			So(f.isDomain("host"), ShouldEqual, true)
			So(f.isIP("host"), ShouldEqual, false)
		})

		Convey("`123.test` should be domain and not IP", func() {
			So(f.isDomain("123.test"), ShouldEqual, true)
			So(f.isIP("123.test"), ShouldEqual, false)
		})

	})
}
