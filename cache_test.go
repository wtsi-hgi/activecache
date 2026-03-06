/*******************************************************************************
 * Copyright (c) 2026 Genome Research Ltd.
 *
 * Author: Michael Woolnough <mw31@sanger.ac.uk>
 *
 * Permission is hereby granted, free of charge, to any person obtaining
 * a copy of this software and associated documentation files (the
 * "Software"), to deal in the Software without restriction, including
 * without limitation the rights to use, copy, modify, merge, publish,
 * distribute, sublicense, and/or sell copies of the Software, and to
 * permit persons to whom the Software is furnished to do so, subject to
 * the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
 * EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
 * MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
 * CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
 * TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 ******************************************************************************/

package activecache

import (
	"io"
	"strconv"
	"testing"
	"testing/synctest"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCache(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		Convey("With a cache", t, func() {
			v := 0
			eFail := false

			c := New(time.Hour, func(k string) (string, error) {
				v++

				if v == 3 {
					return "", io.EOF
				} else if k == "d" {
					return k + strconv.Itoa(v), io.ErrUnexpectedEOF
				} else if eFail && k == "e" {
					return k + strconv.Itoa(v), io.ErrNoProgress
				}

				return k + strconv.Itoa(v), nil
			})

			Reset(c.Stop)

			Convey("You can get non-cached items", func() {
				av, err := c.Get("a")
				So(err, ShouldBeNil)
				So(av, ShouldEqual, "a1")

				bv, err := c.Get("b")
				So(err, ShouldBeNil)
				So(bv, ShouldEqual, "b2")

				cv, err := c.Get("c")
				So(err, ShouldEqual, io.EOF)
				So(cv, ShouldEqual, "")

				cd, err := c.Get("d")
				So(err, ShouldEqual, io.ErrUnexpectedEOF)
				So(cd, ShouldEqual, "d4")

				ce, err := c.Get("e")
				So(err, ShouldEqual, nil)
				So(ce, ShouldEqual, "e5")

				Convey("Removing an item will cause it to be re-retrieved upon next Get", func() {
					c.Remove("a")

					v, err := c.Get("a")
					So(err, ShouldBeNil)
					So(v, ShouldStartWith, "a")
					So(v, ShouldNotEqual, av)
				})

				Convey("Re-retrieving the now cached items will get the same results", func() {
					av, err := c.Get("a")
					So(err, ShouldBeNil)
					So(av, ShouldEqual, "a1")

					bv, err := c.Get("b")
					So(err, ShouldBeNil)
					So(bv, ShouldEqual, "b2")

					cv, err := c.Get("c")
					So(err, ShouldEqual, io.EOF)
					So(cv, ShouldEqual, "")

					cd, err := c.Get("d")
					So(err, ShouldEqual, io.ErrUnexpectedEOF)
					So(cd, ShouldEqual, "d4")
				})

				Convey("and the cache will update the retrieved items after the wait period", func() {
					eFail = true
					time.Sleep(time.Hour + time.Minute)

					v, err := c.Get("a")
					So(err, ShouldBeNil)
					So(v, ShouldStartWith, "a")
					So(v, ShouldNotEqual, av)

					v, err = c.Get("b")
					So(err, ShouldBeNil)
					So(v, ShouldStartWith, "b")
					So(v, ShouldNotEqual, bv)

					v, err = c.Get("c")
					So(err, ShouldBeNil)
					So(v, ShouldStartWith, "c")

					v, err = c.Get("d")
					So(err, ShouldEqual, io.ErrUnexpectedEOF)
					So(v, ShouldStartWith, "d")
					So(v, ShouldNotEqual, cd)

					Convey("items with errors are not updated", func() {
						v, err := c.Get("e")
						So(err, ShouldEqual, nil)
						So(v, ShouldEqual, "e5")

						Convey("but are updated when there's no longer an error", func() {
							eFail = false
							time.Sleep(time.Hour + time.Minute)

							v, err := c.Get("e")
							So(err, ShouldEqual, nil)
							So(v, ShouldStartWith, "e")
							So(v, ShouldNotEqual, ce)
						})
					})
				})
			})
		})
	})
}
