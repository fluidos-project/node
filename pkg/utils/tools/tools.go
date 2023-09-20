package tools

import (
	"time"

	"k8s.io/klog/v2"
)

// GetTimeNow returns the current time in RFC3339 format
func GetTimeNow() string {
	return time.Now().Format(time.RFC3339)
}

// CheckExpiration checks if the timestamp has expired
func CheckExpiration(timestamp string, expTime time.Duration) bool {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		klog.Errorf("Error parsing the transaction start time: %s", err)
		return false
	}
	return time.Since(t) > expTime
}
