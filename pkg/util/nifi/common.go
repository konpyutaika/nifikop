package nifi

import "time"

const (
	// AllNodeServiceTemplate template for Nifi all nodes service
	AllNodeServiceTemplate = "%s-all-node"
	// HeadlessServiceTemplate template for Nifi headless service
	HeadlessServiceTemplate = "%s-headless"
)

// ParseTimeStampToUnixTime parses the given CC timeStamp to time format
func ParseTimeStampToUnixTime(timestamp string) (time.Time, error) {
	timeStampLayout := "Mon, 2 Jan 2006 15:04:05 GMT"
	t, err := time.Parse(timeStampLayout, timestamp)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}
