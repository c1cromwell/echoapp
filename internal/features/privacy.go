package features

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// RetainMessageMetadata reports whether the relay keeps who-sent-what-to-whom records for
// queued messages (sender, conversation, the delivered envelope).
//
// Default false: the messaging gateway records the least it can (docs/RELAY_METADATA_REVIEW.md).
// A row is written only for a message that could not be delivered live; it names the recipient
// alone, and on delivery its content is erased. Set ECHO_RETAIN_MESSAGE_METADATA=true only for a
// deployment whose purpose is retention (Echo Comply's eDiscovery reads these records).
func RetainMessageMetadata() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("ECHO_RETAIN_MESSAGE_METADATA")), "true")
}

// OfflineRetention is how long an undelivered message waits for its recipient before it is
// deleted. ECHO_OFFLINE_RETENTION_DAYS, 1 to 30; default 14.
func OfflineRetention() time.Duration {
	days := 14
	if v, err := strconv.Atoi(strings.TrimSpace(os.Getenv("ECHO_OFFLINE_RETENTION_DAYS"))); err == nil && v >= 1 && v <= 30 {
		days = v
	}
	return time.Duration(days) * 24 * time.Hour
}

// DeliveredRetention is how long the stub of a delivered message is kept (it holds the recipient
// and delivery times only, for delivery and read status). After that the row is deleted.
const DeliveredRetention = 24 * time.Hour
