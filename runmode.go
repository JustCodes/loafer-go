package loafergo

// Mode defines how messages should be dispatched to workers.
type Mode int

const (
	// Parallel processes messages independently without considering group identifiers.
	Parallel Mode = iota

	// PerGroupID ensures startWorker processes messages based on MessageGroupId and custom grouping fields.
	PerGroupID
)
