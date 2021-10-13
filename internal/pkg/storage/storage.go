package storage

type QuorumStorage interface {
	Init(path string) error
	Close() error
	Set(key []byte, val []byte) error
	Delete(key []byte) error
	Get(key []byte) ([]byte, error)
	PrefixForeach(prefix []byte, fn func([]byte, []byte, error) error) error
	Foreach(fn func([]byte, []byte, error) error) error
}
