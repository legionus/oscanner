package types

type TargetReceiver interface {
	Fetch() error
	Snapshot() error
	Cleanup() error
}

// Scanner is common interface for all scanners
type Scanner interface {
	Run(string) error
}
