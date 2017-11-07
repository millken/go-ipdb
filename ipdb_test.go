package ipdb

import (
	"math/rand"
	"testing"
	"time"
)

const data = "ipdb.dat"

var (
	db  *DB
	err error
)

func TestHeader(t *testing.T) {
	if db, err = Load(data); err != nil {
		t.Fatal("Init failed:", err)
	}
	t.Logf("db Header = %v", db.Head)
	ip := "42.63.123.32"
	result, err := db.Find(ip)
	if err == nil {
		t.Logf("find %s => %+v", ip, result)
	}
}

//-----------------------------------------------------------------------------

// Benchmark command
//	go test -bench=Find
//	BenchmarkFind 1000000       1440 ns/op
func BenchmarkFind(b *testing.B) {
	b.StopTimer()
	if db, err = Load(data); err != nil {
		b.Fatal("Init failed:", err)
	}
	rand.Seed(time.Now().UnixNano())
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		rndip := uint32(708803360)
		if _, err := db.FindByUint(rndip); err != nil {
			b.Fatalf("FindByUint %d[%s]: %s", rndip, Long2Ip(rndip).To4(), err.Error())
		}
	}
}
