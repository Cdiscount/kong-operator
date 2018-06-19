package utils

import (
	"strconv"
	"time"
)

//UnixTimeStr use to convert timestamp to humanreadable
func UnixTimeStr(timeStamp int) string {
	i, err := strconv.ParseInt(strconv.Itoa(timeStamp), 10, 64)
	if err != nil {
		panic(err)
	}
	tm := time.Unix(i, 0)
	return tm.String()
}
