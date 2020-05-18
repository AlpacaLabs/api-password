package service

import (
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	"encoding/json"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_GetHashConfiguration(t *testing.T) {
	Convey("Test GetHashConfiguration", t, func() {
		c := GetHashConfiguration(time.Millisecond * 500)
		b, err := json.Marshal(c)
		So(err, ShouldBeNil)
		log.Info(string(b))
	})
}
