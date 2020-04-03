package util

import "time"

func GetCurrentUnixNano() int64 {
	return time.Now().UnixNano()
}
