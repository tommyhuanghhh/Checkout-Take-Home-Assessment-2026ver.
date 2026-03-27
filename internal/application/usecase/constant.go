package usecase
import (
	"time"
)

const (
	IdempotencyInProgressMarker = "IN_PROGRESS"
	IdempotencyTTL              = 24 * time.Hour
)