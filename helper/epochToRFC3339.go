package common

import "time"

func EpochToRFC3339(epoch int64) string {
	return time.Unix(epoch, 0).UTC().Format(time.RFC3339)
}
