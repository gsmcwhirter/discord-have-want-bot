package storage

//go:generate protoc --go_out=. --proto_path=. ./trialapi.proto

// TrialAPI TODOC
type TrialAPI interface {
	NewTransaction(writable bool) (TrialAPITx, error)
}

// TrialAPITx TODOC
type TrialAPITx interface {
	Commit() error
	Rollback() error

	GetTrial(name string) (Trial, error)
	AddTrial(name string) (Trial, error)
	SaveTrial(trial Trial) error
}

// Trial TODOC
type Trial interface {
	GetName() string

	SetName(name string)

	Serialize() ([]byte, error)
}
