package ulid_test

import (
	"encoding/binary"
	"encoding/hex"
	"math"
	"strings"
	"testing"

	"github.com/filecoin-project/go-address"
	"github.com/glifio/graph/internal/ulid"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestUlidWithNonce(t *testing.T) {
	var height uint64 = 1
	addr := address.NewForTestGetter()
	baddrs := addr().Bytes()
	t.Log(addr().String())
	t.Log(hex.EncodeToString(baddrs))
	t.Log(hex.EncodeToString(baddrs[len(baddrs)-2:]))

	id1, _ := ulid.NewReverse(height)
	entropy1 := id1.Entropy()
	binary.BigEndian.PutUint64(entropy1, math.MaxUint64-100)
	entropy1[8] = baddrs[len(baddrs)-2]
	entropy1[9] = baddrs[len(baddrs)-1]
	id1.SetEntropy(entropy1)

	// entropy1 := node.ZeroReader{Nonce: math.MaxUint64 - 100}
	// entropy2 := node.ZeroReader{Nonce: math.MaxUint64 - 200}
	//entropy := node.ZeroReader{Nonce: 1500}

	id2, _ := ulid.NewReverse(height)
	entropy2 := id2.Entropy()
	binary.BigEndian.PutUint64(entropy2, math.MaxUint64-200)
	if len(baddrs) > 2 {
		copy(entropy2[8:], baddrs[len(baddrs)-2:])
	}
	id2.SetEntropy(entropy2)
	t.Log(hex.EncodeToString(id1[:]))
	t.Log(hex.EncodeToString(id2[:]))
	t.Log(id1.String())
	t.Log(id2.String())
	if strings.Compare(id1.String(), id2.String()) != 1 {
		t.Fatalf(`ulid = %s should come before %s`, id2.String(), id1.String())
	}
}
