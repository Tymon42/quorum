package storage

import (
	badger "github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/options"
)

type QSBadger struct {
	db *badger.DB
}

var DefaultLogFileSize int64 = 16 << 20
var DefaultMemTableSize int64 = 8 << 20
var DefaultMaxEntries uint32 = 50000
var DefaultBlockCacheSize int64 = 32 << 20
var DefaultCompressionType = options.Snappy
var DefaultPrefetchSize = 10

func (s *QSBadger) Init(path string) error {
	var err error
	s.db, err = badger.Open(badger.DefaultOptions(path).WithValueLogFileSize(DefaultLogFileSize).WithMemTableSize(DefaultMemTableSize).WithValueLogMaxEntries(DefaultMaxEntries).WithBlockCacheSize(DefaultBlockCacheSize).WithCompression(DefaultCompressionType).WithLoggingLevel(badger.ERROR))
	if err != nil {
		return err
	}
	return nil
}

func (s *QSBadger) Close() error {
	return s.db.Close()
}

func (s *QSBadger) Set(key []byte, val []byte) error {
	return s.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(key, val)
		err := txn.SetEntry(e)
		return err
	})
}

func (s *QSBadger) Delete(key []byte) error {
	return s.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(key))
		return err
	})

}

func (s *QSBadger) Get(key []byte) ([]byte, error) {
	var val []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		val, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return nil
	})
	return val, err
}

func (s *QSBadger) PrefixForeach(prefix []byte, fn func([]byte, []byte, error) error) error {
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = DefaultPrefetchSize
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.KeyCopy(nil)
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			ferr := fn(key, val, nil)
			if ferr != nil {
				return ferr
			}
		}
		return nil
	})
	return err
}

func (s *QSBadger) Foreach(fn func([]byte, []byte, error) error) error {
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = DefaultPrefetchSize
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.KeyCopy(nil)
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			ferr := fn(key, val, nil)
			if ferr != nil {
				return ferr
			}
		}
		return nil
	})
	return err
}
