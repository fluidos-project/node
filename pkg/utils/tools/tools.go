// Copyright 2022-2024 FLUIDOS Project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tools

import (
	"time"

	"k8s.io/klog/v2"
)

// GetTimeNow returns the current time in RFC3339 format.
func GetTimeNow() string {
	return time.Now().Format(time.RFC3339)
}

// GetExpirationTime returns the current time plus h hours, m minutes and s seconds in RFC3339 format.
func GetExpirationTime(h, m, s int) string {
	return time.Now().Add(time.Hour * time.Duration(h)).Add(time.Minute * time.Duration(m)).Add(time.Second * time.Duration(s)).Format(time.RFC3339)
}

// CheckExpiration checks if the expirationTimestamp has already expired.
func CheckExpiration(expirationTimestamp string) bool {
	t, err := time.Parse(time.RFC3339, expirationTimestamp)
	if err != nil {
		klog.Errorf("Error parsing the transaction start time: %s", err)
		return false
	}
	return time.Now().After(t)
}

// CheckExpirationSinceTime checks if the expirationTimestamp has already expired.
func CheckExpirationSinceTime(sinceTimestamp string, expirationLapse time.Duration) bool {
	t, err := time.Parse(time.RFC3339, sinceTimestamp)
	if err != nil {
		klog.Errorf("Error parsing the transaction start time: %s", err)
		return false
	}
	return time.Since(t) > expirationLapse
}
