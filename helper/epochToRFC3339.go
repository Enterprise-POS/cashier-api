package common

import "time"

func EpochToRFC3339(epoch int64) string {
	t := time.Unix(epoch, 0).UTC()
	// PostgreSQL timestamptz expects ISO 8601 format
	return t.Format(time.RFC3339)
}
