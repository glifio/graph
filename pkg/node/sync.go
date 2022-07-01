package node

import (
	"context"
	"fmt"
	"log"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/glifio/graph/pkg/kvdb"
	"github.com/glifio/graph/pkg/postgres"
	"github.com/spf13/viper"
)

var syncRunning bool

func (t *Node) SyncTimerStart(confidence uint32) {
	t.ticker = time.NewTicker(30 * time.Second)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-t.ticker.C:
				SyncTipSetTop(context.TODO())
			}
		}
	}()
}

func (t *Node) SyncTimerStop() {
	if t.ticker != nil {
		t.ticker.Stop()
		fmt.Println("Ticker stopped")
	}
}

func GetMaxHeight() uint32 {
	db := kvdb.Open()
	defer db.Discard()

	ch, _ := lotus.api.ChainHead(context.Background())
	height := uint32(ch.Height()) - viper.GetUint32("confidence")
	limit := height - viper.GetUint32("confidence")
	for {
		// get messages in tipset from cache
		key := []byte(fmt.Sprintf("h:%d", height))

		if db.Exists(key) {
			return height
		}

		height--

		// don't go crazy
		if height <= limit {
			return height
		}
	}
}

func Sync(p0 context.Context, _confidence uint64, _height uint64, _length uint64) {
	var err error
	var success, fail, count uint

	if syncRunning {
		return
	}

	syncRunning = true

	// if true {
	// 	f, err := os.Create("cpuprofile")
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	pprof.StartCPUProfile(f)
	// 	defer pprof.StopCPUProfile()
	// }

	db := kvdb.Open()
	defer db.Discard()

	// backfill the cache
	height := _height
	length := _length

	if _height == 0 {
		ch, _ := lotus.api.ChainHead(context.Background())
		height = uint64(ch.Height()) - _confidence
	}
	if _length == 0 {
		length = height
	}
	gen, _ := lotus.api.ChainGetGenesis(context.Background())
	log.Println("genesis: ", gen.Key())

	//db.DB().RunValueLogGC(0.5)

	log.Printf("sync -> tipset %d len %d\n", abi.ChainEpoch(height), length)
	ts, _ := lotus.api.ChainGetTipSetByHeight(context.Background(), abi.ChainEpoch(height), types.EmptyTSK)
	tsk := ts.Key()

	start := time.Now()

	for {
		// start multi txn
		//db.Multi()
		db.Batch()
		defer db.Cancel()

		// get tipset from cache
		if ts, err = GetTipSet(tsk); err != nil {
			log.Println(err)
			syncRunning = false
			return
		}

		_, err = SyncTipSetMessages(p0, ts)
		if err != nil {
			fail++
		} else {
			success++
		}

		if ts.Height()%100 == 0 {
			timeElapsed := time.Since(start)
			estimate := (float64(ts.Height()) / 100) * timeElapsed.Seconds()
			dur := time.Duration(estimate) * time.Second
			start = time.Now()

			log.Printf("sync -> stats h:%s success:%d fail:%d dur:%s est:%s\n", ts.Height(), success, fail, timeElapsed, dur)
			success = 0
			fail = 0
		}

		// execute commit
		//_ = db.Exec()
		db.Flush()
		db.Cancel()

		// genesis we're done
		if ts.Height() == 0 {
			log.Printf("sync -> genesis\n")
			break
		}

		count++
		if count >= uint(length) {
			log.Printf("sync -> done\n")
			break
		}

		tsk = ts.Parents()
	}

	//_ = db.Exec()
	db.Flush()

	syncRunning = false

	lsm, vlog := db.DB().Size()
	log.Printf("sync -> commit %d %d\n", lsm/1048576, vlog/1048576)
}

func SyncLily(p0 context.Context, _height uint64, _length uint64) error {
	var err error
	var success, fail, count uint

	db := kvdb.Open()
	defer db.Discard()

	log.Printf("sync -> tipset %d\n", abi.ChainEpoch(_height))

	// backfill the cache from lily db
	ts, err := lotus.api.ChainGetTipSetByHeight(p0, abi.ChainEpoch(_height), types.EmptyTSK)
	if err != nil {
		log.Println(err)
		return err
	}

	tsk := ts.Key()

	start := time.Now()

	for {
		// start multi txn
		//db.Multi()
		db.Batch()
		defer db.Cancel()

		// get messages in tipset from cache
		if ts, err = GetTipSet(tsk); err != nil {
			log.Println(err)
			return err
		}

		err = SyncMessagesFromLily(p0, ts)
		if err != nil {
			fail++
		} else {
			success++
		}

		if ts.Height()%100 == 0 {
			timeElapsed := time.Since(start)
			estimate := (float64(ts.Height()) / 100) * timeElapsed.Seconds()
			dur := time.Duration(estimate) * time.Second
			start = time.Now()

			log.Printf("sync -> stats h:%s success:%d fail:%d dur:%s est:%s\n", ts.Height(), success, fail, timeElapsed, dur)
		}

		// genesis we're done
		if ts.Height() == 0 {
			log.Printf("sync -> genesis\n")
			break
		}

		// execute commit
		// _ = db.Exec()
		db.Flush()
		db.Cancel()

		count++
		if count >= uint(_length) {
			log.Printf("sync -> done\n")
			break
		}

		tsk = ts.Parents()
	}

	// _ = db.Exec()
	db.Flush()
	db.Cancel()

	lsm, vlog := db.DB().Size()
	log.Printf("sync -> commit %d %d\n", lsm/1048576, vlog/1048576)
	return nil
}

func SyncTipSetTop(p0 context.Context) error {
	ch, err := lotus.api.ChainHead(context.Background())
	if err != nil {
		return err
	}

	height := uint32(ch.Height()) - viper.GetViper().GetUint32("confidence") // todo unsafe
	ts, err := lotus.api.ChainGetTipSetByHeight(context.Background(), abi.ChainEpoch(height), types.EmptyTSK)
	if err != nil {
		return err
	}

	kvdb.Open().Batch()
	defer kvdb.Open().Cancel()

	_tsk := ts.Key()

	loop := 0
	for {
		if ExistsTipSet(_tsk) {
			break
		}

		ts, err := GetTipSet(_tsk)
		if err != nil {
			log.Println(err)
			return err
		}

		count, err := SyncTipSetMessages(p0, ts)
		if err != nil {
			return err
		}
		log.Printf("sync -> timer h:%d messages:%d\n", ts.Height(), count)

		loop++
		if loop > 5 {
			break
		}

		_tsk = ts.Parents()
	}

	kvdb.Open().Flush()
	return nil
}

func SyncMessagesFromLily(p0 context.Context, ts *types.TipSet) error {

	res, err := postgres.GetMessagesInTipset(uint(ts.Height()))

	if err != nil {
		log.Printf("error tipset: h:%d\n %s\n", ts.Height(), err)
		return err
	}

	for _, msg := range res {
		// store messages
		SetMessage(msg, ts)
	}

	// store list of tipset messages
	SetTipSetMessages(ts.Key(), res)

	// log.Printf("sync -> tipset msg save h:%d m:%d\n", height, len(cids))

	return nil
}

func SyncTipSetMessages(p0 context.Context, ts *types.TipSet) (int, error) {

	res, err := GetTipSetMessages(ts)

	if err != nil {
		log.Printf("error tipset: h:%d\n %s\n", ts.Height(), err)
		return 0, err
	}

	//compute, err := t.api.StateCompute(p0, api.LookbackNoLimit, tmsg, ts.Key())
	//compute.Trace[0].GasCost.GasUsed

	return len(res), nil
}

func ValidateMessages(height uint64) {
	db := kvdb.Open()
	i := 0

	log.Printf("validate -> messages")

	prefix := []byte("cid:")

	_ = db.DB().View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		//opts.Reverse = true
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			_ = item.Value(func(v []byte) error {
				msg, err := DecodeMessage(v)
				if err == nil {
					log.Printf("validate -> hit %s", msg.Cid)
					// AddAddressToMessageIndex(msg)
				}
				return nil
			})
			i++
			if i > 10 {
				return nil
			}
		}
		return nil
	})
	log.Printf("search counter: %d\n", i)
}
