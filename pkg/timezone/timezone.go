package timezone

import "time"

const Name = "Asia/Jakarta"

var Loc *time.Location

func init() {
	loc, err := time.LoadLocation(Name)
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	Loc = loc
}
