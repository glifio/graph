package node

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/spf13/viper"
)

var once1 sync.Once

type Cache struct {
	cache *ristretto.Cache
}

// variabel Global
var cache *Cache

func GetCacheInstance() *Cache {

	once1.Do(func() {
		ris, err := ristretto.NewCache(&ristretto.Config{
			NumCounters: 1e7,     // number of keys to track frequency of (10M).
			MaxCost:     1 << 30, // maximum cost of cache (1GB).
			BufferItems: 64,      // number of keys per Get buffer.
		})
		if err != nil {
			log.Fatal(err)
		}

		cache = &Cache{
			cache: ris,
		}
	})

	return cache
}

func (c *Cache) Cache() *ristretto.Cache {
	return c.cache
}

func (c *Cache) Close() {
	c.cache.Close()
}

func (t *Node) StartCache() {
	log.Println("cache -> init")
	ctx, cancel := context.WithCancel(context.Background())

	// listen for new chainhead
	go func() {
		for {
			for {
				log.Printf("cache -> subscribe to chainhead\n")
				chain, err := lotus.api.ChainNotify(context.Background())

				if err == nil {
					for headchanges := range chain {
						for _, elem := range headchanges {
							switch elem.Type {
							case "current":
								fallthrough
							case "apply":
								ctx, cancel = context.WithCancel(context.Background())
								go cacheTipset(ctx, t, elem)
							case "revert":
								// log.Printf("cache -> tipset %s %s\n", elem.Val.Height(), elem.Type)
								cancel()
							}
						}
					}
				}
				log.Printf("cache -> subscription failed: %s\n", err)
				time.Sleep(15 * time.Second)
			}
		}
	}()

	// backfill the cache
	log.Printf("cache -> backfill\n")
	ts, _ := lotus.api.ChainHead(context.Background())
	for i := 0; i < viper.GetInt("confidence"); i++ {
		go func(tipset types.TipSet, i int) {
			//log.Printf("cache -> backfill tipset %s %d\n", ts.Height(), i)
			t.ChainGetMessagesInTipset(context.Background(), ts.Key(), i)
		}(*ts, i)
		ts, _ = t.ChainGetTipSet(context.Background(), ts.Parents())
	}
	log.Printf("cache -> operational\n")
}

func cacheTipset(ctx context.Context, t *Node, elem *lotusapi.HeadChange) {
	select {
	case <-time.After(2000 * time.Millisecond):
		//log.Printf("cache -> tipset %s %s\n", elem.Val.Height(), elem.Type)
		t.ChainGetMessagesInTipset(context.Background(), elem.Val.Key(), int(elem.Val.Height()))
		GetCacheInstance().cache.Set("node/chainhead/tipsetkey", elem.Val.Key(), 1)
	case <-ctx.Done():
		// log.Printf("cache -> tipset %s %s\n", elem.Val.Height(), "halted")
	}
}
