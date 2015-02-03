package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHostDomainAndIP(t *testing.T) {
	Convey("Test Host File Domain and IP regex", t, func() {
		Convey("1.1.1.1 should be IP and not domain", func() {
			So(isIP("1.1.1.1"), ShouldEqual, true)
			So(isDomain("1.1.1.1"), ShouldEqual, false)
		})

		Convey("2001:470:20::2 should be IP and not domain", func() {
			So(isIP("2001:470:20::2"), ShouldEqual, true)
			So(isDomain("2001:470:20::2"), ShouldEqual, false)
		})

		Convey("`host` should be domain and not IP", func() {
			So(isDomain("host"), ShouldEqual, true)
			So(isIP("host"), ShouldEqual, false)
		})

		Convey("`123.test` should be domain and not IP", func() {
			So(isDomain("123.test"), ShouldEqual, true)
			So(isIP("123.test"), ShouldEqual, false)
		})

	})
}
