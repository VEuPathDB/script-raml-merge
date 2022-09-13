package script_test

import (
	"testing"

	"script-raml-merger/internal/script"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTypeToFiles_Append(t *testing.T) {
	Convey("TypeNameToParentFileMap.Append", t, func() {
		tmp := make(script.TypeNameToParentFileMap, 10)
		tmp.Append("foo", "bar")

		v, ok := tmp["foo"]
		So(ok, ShouldBeTrue)
		So(v, ShouldNotBeNil)
		So(len(v), ShouldEqual, 1)
		So(v["bar"], ShouldBeTrue)
	})
}
