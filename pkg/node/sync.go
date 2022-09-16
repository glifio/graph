package node

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/glifio/graph/gql/model"
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
	var total int

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

	// start a write batch
	wb := db.Batch()
	wb.SetMaxPendingTxns(512)
	defer wb.Cancel()

	for {
		// get tipset from cache
		if ts, err = GetTipSet(tsk, wb); err != nil {
			log.Println(err)
			syncRunning = false
			return
		}

		c, err := SyncTipSetMessages(p0, ts, wb)
		if err != nil {
			fail++
		} else {
			success++
		}
		total += c

		if ts.Height()%100 == 0 {
			timeElapsed := time.Since(start)
			estimate := (float64(ts.Height()) / 100) * timeElapsed.Seconds()
			dur := time.Duration(estimate) * time.Second
			start = time.Now()

			log.Printf("sync -> stats h:%s success:%d fail:%d dur:%s est:%s\n", ts.Height(), success, fail, timeElapsed, dur)
			success = 0
			fail = 0
		}

		//		log.Printf("sync -> h:%d messages:%d\n", ts.Height(), total)

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

	wb.Flush()

	syncRunning = false

	lsm, vlog := db.DB().Size()
	log.Printf("sync -> commit %d %d\n", lsm/1048576, vlog/1048576)
}

// ctx, cancel = context.WithCancel(context.Background())
// go SyncTipsetStart(ctx)

var jobs chan types.TipSetKey
var results chan types.TipSetKey
var status chan *model.Status

var currentStatus = model.Status{Height: 0, Estimate: 0}

func SyncStatus() *model.Status {
	return &currentStatus
}

func SyncTipsetStart(ctx context.Context, _confidence uint64, _height uint64, _length uint64) {
	ch, _ := lotus.api.ChainHead(ctx)
	height := uint64(ch.Height()) - _confidence
	ts, _ := lotus.api.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(height), types.EmptyTSK)
	tsk := ts.Key()

	jobs = make(chan types.TipSetKey, 5)
	results = make(chan types.TipSetKey, 5)
	go SyncTipsetWorker(ctx, 1, jobs, results)

	log.Printf("tipset -> start at %d\n", ts.Height())

	// if true {
	// 	f, err := os.Create("cpuprofile")
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	pprof.StartCPUProfile(f)
	// 	defer pprof.StopCPUProfile()
	// }

	jobs <- tsk
	for {
		select {
		case tsk = <-results:
			if tsk == types.EmptyTSK {
				return
			}
			jobs <- tsk
		case <-ctx.Done():
			time.Sleep(3 * time.Second)
			close(jobs)
			close(results)
			log.Printf("tipset -> %s\n", "halted")
			return
		}
	}
}

func SyncTipsetWorker(ctx context.Context, id int, jobs <-chan types.TipSetKey, result chan<- types.TipSetKey) {
	db := kvdb.Open()

	// start a write batch
	wb := db.Batch()
	wb.SetMaxPendingTxns(64)
	defer wb.Cancel()

	start := time.Now()

	for {
		select {
		case key := <-jobs:
			// get tipset
			ts, err := GetTipSet(key, wb)
			if err != nil {
				log.Println(err)
				result <- types.TipSetKey{}
			}
			// get messages
			_, _ = UpdateTipSetMessages(ts, wb)

			// log stats
			if ts.Height()%100 == 0 {
				timeElapsed := time.Since(start)
				estimate := (float64(ts.Height()) / 100) * timeElapsed.Seconds()
				dur := time.Duration(estimate) * time.Second
				start = time.Now()
				currentStatus.Height = int64(ts.Height())
				currentStatus.Estimate = int64(estimate)
				log.Printf("tipset(%d) -> stats h:%s dur:%s est:%s\n", id, ts.Height(), timeElapsed, dur)
			}
			// return parent tipset
			if ts.Height() == 0 {
				log.Printf("tipset(%d) genesis -> parent:%s\n", id, ts.Parents())
				result <- types.EmptyTSK
			} else {
				result <- ts.Parents()
			}
		case <-ctx.Done():
			log.Printf("tipset worker -> %s\n", "halted")
			wb.Flush()
			return
		}
	}
}

func SyncTipset_deleteme(p0 context.Context, _confidence uint64, _height uint64, _length uint64) {
	var count uint

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

	//db.DB().RunValueLogGC(0.5)

	log.Printf("sync -> tipset %d len %d\n", abi.ChainEpoch(height), length)
	ts, _ := lotus.api.ChainGetTipSetByHeight(p0, abi.ChainEpoch(height), types.EmptyTSK)
	tsk := ts.Key()

	start := time.Now()

	// start a write batch
	wb := db.Batch()
	wb.SetMaxPendingTxns(512)
	defer wb.Cancel()

	for {
		select {
		case <-p0.Done():
			fmt.Println("halted operation2")
			wb.Flush()
			return
		default:
			// get tipset from cache
			ts, err := GetTipSet(tsk, wb)
			if err != nil {
				log.Println(err)
				return
			}

			if ts.Height()%1000 == 0 {
				timeElapsed := time.Since(start)
				estimate := (float64(ts.Height()) / 1000) * timeElapsed.Seconds()
				dur := time.Duration(estimate) * time.Second
				start = time.Now()
				log.Printf("tipset sync -> stats h:%s dur:%s est:%s\n", ts.Height(), timeElapsed, dur)
			}

			// genesis we're done
			if ts.Height() == 0 {
				log.Printf("tipset sync -> genesis\n")
				break
			}

			count++
			if count >= uint(length) {
				log.Printf("tipset sync -> done\n")
				break
			}

			tsk = ts.Parents()
		}
	}
	log.Printf("tipset sync -> commit\n")
}

func SyncMessages(p0 context.Context, _confidence uint64, _height uint64, _length uint64) {
	log.Printf("sync -> messages")

	var i uint

	db := kvdb.Open()
	defer db.Discard()

	prefix := []byte("t:")

	wb := kvdb.Open().Batch()
	defer wb.Cancel()

	_ = db.DB().View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			_ = item.Value(func(val []byte) error {
				ts := &types.TipSet{}
				if err := ts.UnmarshalCBOR(bytes.NewReader(val)); err == nil {
					_, err = UpdateTipSetMessages(ts, wb)
					//_, err = SyncTipSetMessages(p0, ts, wb)
					if err != nil {
						log.Println(err)
					}
				}
				return nil
			})
			i++

			if i%1000 == 0 {
				log.Printf("message sync -> stats count:%d\n", i)
			}
		}

		wb.Flush()
		return nil
	})
	log.Printf("message sync counter: %d\n", i)

	log.Printf("message sync -> commit\n")
}

// func SyncLily(p0 context.Context, _height uint64, _length uint64) error {
// 	var err error
// 	var success, fail, count uint

// 	db := kvdb.Open()
// 	defer db.Discard()

// 	log.Printf("sync -> tipset %d\n", abi.ChainEpoch(_height))

// 	// backfill the cache from lily db
// 	ts, err := lotus.api.ChainGetTipSetByHeight(p0, abi.ChainEpoch(_height), types.EmptyTSK)
// 	if err != nil {
// 		log.Println(err)
// 		return err
// 	}

// 	tsk := ts.Key()

// 	start := time.Now()

// 	for {
// 		// start multi txn
// 		//db.Multi()
// 		db.Batch()
// 		defer db.Cancel()

// 		// get messages in tipset from cache
// 		if ts, err = GetTipSet(tsk); err != nil {
// 			log.Println(err)
// 			return err
// 		}

// 		err = SyncMessagesFromLily(p0, ts)
// 		if err != nil {
// 			fail++
// 		} else {
// 			success++
// 		}

// 		if ts.Height()%100 == 0 {
// 			timeElapsed := time.Since(start)
// 			estimate := (float64(ts.Height()) / 100) * timeElapsed.Seconds()
// 			dur := time.Duration(estimate) * time.Second
// 			start = time.Now()

// 			log.Printf("sync -> stats h:%s success:%d fail:%d dur:%s est:%s\n", ts.Height(), success, fail, timeElapsed, dur)
// 		}

// 		// genesis we're done
// 		if ts.Height() == 0 {
// 			log.Printf("sync -> genesis\n")
// 			break
// 		}

// 		// execute commit
// 		// _ = db.Exec()
// 		db.Flush()
// 		db.Cancel()

// 		count++
// 		if count >= uint(_length) {
// 			log.Printf("sync -> done\n")
// 			break
// 		}

// 		tsk = ts.Parents()
// 	}

// 	// _ = db.Exec()
// 	db.Flush()
// 	db.Cancel()

// 	lsm, vlog := db.DB().Size()
// 	log.Printf("sync -> commit %d %d\n", lsm/1048576, vlog/1048576)
// 	return nil
// }

func SyncTipSetTop(p0 context.Context) error {
	db := kvdb.Open()

	ch, err := lotus.api.ChainHead(context.Background())
	if err != nil {
		return err
	}

	height := uint32(ch.Height()) - viper.GetViper().GetUint32("confidence") // todo unsafe
	ts, err := lotus.api.ChainGetTipSetByHeight(context.Background(), abi.ChainEpoch(height), types.EmptyTSK)
	if err != nil {
		return err
	}

	wb := db.Batch()
	wb.SetMaxPendingTxns(512)
	defer wb.Cancel()

	_tsk := ts.Key()

	loop := 0
	for {
		if ExistsTipSet(_tsk) {
			break
		}

		ts, err := GetTipSet(_tsk, wb)
		if err != nil {
			log.Println(err)
			return err
		}

		count, err := SyncTipSetMessages(p0, ts, wb)
		if err != nil {
			return err
		}
		log.Printf("sync -> timer h:%d messages:%d\n", ts.Height(), count)

		loop++
		if loop > 10 {
			break
		}

		_tsk = ts.Parents()
	}

	wb.Flush()
	return nil
}

func SyncMessagesFromLily(p0 context.Context, ts *types.TipSet, wb *badger.WriteBatch) error {

	res, err := postgres.GetMessagesInTipset(uint(ts.Height()))

	if err != nil {
		log.Printf("error tipset: h:%d\n %s\n", ts.Height(), err)
		return err
	}

	for _, msg := range res {
		// store messages
		SetMessage(msg, ts, wb)
	}

	// store list of tipset messages
	SetTipSetMessages(ts.Key(), res, wb)

	// log.Printf("sync -> tipset msg save h:%d m:%d\n", height, len(cids))

	return nil
}

func SyncTipSetMessages(p0 context.Context, ts *types.TipSet, wb *badger.WriteBatch) (int, error) {

	res, err := GetTipSetMessages(ts, wb)

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

		wb := kvdb.Open().Batch()
		defer wb.Cancel()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			_ = item.Value(func(v []byte) error {
				msg, err := DecodeLotusMessage(v)
				if err == nil {
					// log.Printf("validate -> hit %s nonce %d", string(it.Item().Key()), msg.Nonce)
					AddAddressToMessageIndex(context.Background(), msg, &types.TipSet{}, wb)
				}
				return nil
			})
			i++

			if i%10000 == 0 {
				log.Printf("validate sync -> count:%d\n", i)
			}
		}

		wb.Flush()
		return nil
	})
	log.Printf("search counter: %d\n", i)
}
