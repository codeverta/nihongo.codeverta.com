package jobs

import (
	"fmt"
	"strconv"
	"time"
)

func Interval(value string) (time.Duration, error) {
	if value == "" {
		return 168 * time.Hour, nil
	}
	hours, err := strconv.Atoi(value)
	if err != nil || (hours != 168 && hours != 84) {
		return 0, fmt.Errorf("JOBS_SYNC_INTERVAL_HOURS must be 168 (weekly) or 84 (twice weekly)")
	}
	return time.Duration(hours) * time.Hour, nil
}
