package undoer

import (
	"time"
)

type Log interface {
	String() string
	Time() time.Time
}
