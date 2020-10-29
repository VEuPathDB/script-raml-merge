package script_test

import (
	"script-raml-merger/internal/script"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTypeToFiles_Append(t *testing.T) {
	Convey("TypeToFiles.Append", t, func() {
		tmp := make(script.TypeToFiles, 10)
		tmp.Append("foo", "bar")

		v, ok := tmp["foo"]
		So(ok, ShouldBeTrue)
		So(v, ShouldNotBeNil)
		So(len(v), ShouldEqual, 1)
		So(v["bar"], ShouldBeTrue)
	})
}
