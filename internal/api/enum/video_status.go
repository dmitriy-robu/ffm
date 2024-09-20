package enum

type VideoStatus int

const (
	VideoStatusUnknown VideoStatus = iota
	VideoStatusProcessing
	VideoStatusProcessed
	VideoStatusFailed
	VideoStatusDisabled
)

func (s VideoStatus) String() string {
	return [...]string{"unknown", "processing", "processed", "failed"}[s]
}
