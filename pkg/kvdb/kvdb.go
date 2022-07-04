package kvdb

import (
	"errors"
	"log"
	"sync"

	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/options"
	"github.com/spf13/viper"
)

var once sync.Once

type KvDB struct {
	db  *badger.DB
	txn *badger.Txn
	wb  *badger.WriteBatch
}

// variabel Global
var kvdb *KvDB

func Open() *KvDB {

	once.Do(func() {
		path := viper.GetViper().GetString("path")
		opts := badger.DefaultOptions(path)
		opts.MemTableSize = 16 << 20 // 64 << 20,
		opts.NumMemtables = 2        // 5
		opts.BaseTableSize = 16 << 20
		opts.IndexCacheSize = 100 << 20 // 100 mb or some other size based on the amount of data
		opts.Compression = options.ZSTD
		opts.ZSTDCompressionLevel = 3
		opts.BlockCacheSize = 256 << 20
		opts.Logger = nil
		badgerDb, err := badger.Open(opts)
		if err != nil {
			log.Fatal(err)
		}

		kvdb = &KvDB{
			db: badgerDb,
		}
	})

	return kvdb
}

func (kv *KvDB) DB() *badger.DB {
	return kv.db
}

func (kv *KvDB) Close() error {
	return kv.db.Close()
}

// Check if key exists.
func (kv *KvDB) Exists(key []byte) bool {
	err := kv.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		return err
	})
	return err == nil
}

func (kv *KvDB) Multi() *badger.Txn {
	if kv.txn == nil {
		kv.txn = kv.db.NewTransaction(true)
	}
	return kv.txn
}

func (kv *KvDB) Batch() *badger.WriteBatch {
	return kv.db.NewWriteBatch()
}

func (kv *KvDB) Flush() *badger.WriteBatch {
	if kv.wb != nil {
		kv.wb.Flush()
	}
	return kv.wb
}

func (kv *KvDB) Cancel() *badger.WriteBatch {
	if kv.wb != nil {
		kv.wb.Cancel()
		kv.wb = nil
	}
	return kv.wb
}

func (kv *KvDB) Exec() error {
	if kv.txn != nil {
		err := kv.txn.Commit()
		kv.txn = nil
		return err
	}
	return nil
}

func (kv *KvDB) Discard() {
	if kv.txn != nil {
		kv.txn.Discard()
		kv.txn = nil
	}
}

func (kv *KvDB) Get(key []byte) ([]byte, error) {
	var v []byte
	err := kv.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			v = append([]byte{}, val...)
			return nil
		})
		return err
	})
	return v, err
}

func (kv *KvDB) Del(key []byte) error {
	err := kv.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete(key)
		return err
	})
	return err
}

func (kv *KvDB) Set(key []byte, val []byte) error {
	// Write batch
	if kv.wb != nil {
		//log.Printf("kv -> Set wb %s\n", string(key))
		return kv.wb.Set(key, val)
	}

	// Transaction
	if kv.txn != nil {
		//log.Printf("kv -> Set txn %s\n", string(key))
		err := kv.txn.Set(key, val)

		// If txn too big: commit, start a new txn and try again
		if err == badger.ErrTxnTooBig {
			log.Printf("kv -> txn too big\n")
			kv.Exec()
			kv.Multi()
			return kv.txn.Set(key, val)
		}

		return err
	}

	txn := kv.db.NewTransaction(true)
	defer txn.Discard()

	err := txn.Set(key, val)

	if err != nil {
		return err
	}

	return txn.Commit()
}

func (kv *KvDB) SetWb(key []byte, val []byte, wb *badger.WriteBatch) error {
	// Write batch
	if wb == nil {
		return errors.New("error: write batch missing")
	}
	return wb.Set(key, val)
}

func (kv *KvDB) SetTx(key []byte, val []byte, txn *badger.Txn) error {
	// Transaction
	if txn == nil {
		return errors.New("error: transaction missing")
	}
	return txn.Set(key, val)
}

// Set key to hold string value if key does not exist.
// In that case, it is equal to SET. When key already
// holds a value, no operation is performed.
// SETNX is short for "SET if Not eXists".
func (kv *KvDB) SetNX(key []byte, val []byte) error {
	return kv.db.View(func(txn *badger.Txn) error {
		if _, err := txn.Get(key); err == badger.ErrKeyNotFound {
			return kv.Set(key, val)
		}
		return nil
	})
}

func (kv *KvDB) SetNxTx(key []byte, val []byte, txn *badger.Txn) error {
	return kv.db.View(func(txn *badger.Txn) error {
		if _, err := txn.Get(key); err == badger.ErrKeyNotFound {
			return kv.Set(key, val)
		}
		return nil
	})
}

func (kv *KvDB) SetNxWb(key []byte, val []byte, wb *badger.WriteBatch) error {
	return kv.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return kv.Set(key, val)
		}
		return err
	})
}

func (kv *KvDB) SetNXs(key []byte, val []byte) error {
	return kv.db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(key); err == badger.ErrKeyNotFound {
			return txn.Set(key, val)
		}
		return nil
	})
}

func (kv *KvDB) Search(prefix []byte, limit uint, offset uint) ([][]byte, error) {
	res := [][]byte{}
	var i uint = 0
	log.Printf("search -> k:%s l:%d o:%d", string(prefix), limit, offset)
	err := kv.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		//opts.Reverse = true
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); /*&& i < limit+offset*/ it.Next() {
			item := it.Item()
			if i >= offset && i < limit+offset {
				err := item.Value(func(v []byte) error {
					val := make([]byte, len(v))
					log.Printf("search -> hit %d %s", i, item.Key())
					copy(val, v)
					// append makes a copy of value
					res = append(res, val)
					return nil
				})
				if err != nil {
					return err
				}
			}
			i++
		}
		return nil
	})
	log.Printf("search counter: %d\n", i)
	return res, err
}
