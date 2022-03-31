package filequeue_test

import (
	"sync"
	"testing"
	"unsafe"

	"github.com/sunthinker/filequeue"
)

type addTxs struct {
	address [42]byte
	txs     int
}

var add1 addTxs
var add2 addTxs

const (
	str        = "0x81b7e08f65bdf5648606c89998a9cc8164397647"
	maxtestnum = 10000
)

func TestFileWR(t *testing.T) {

	var wg sync.WaitGroup
	//send
	for i := 0; i < maxtestnum/100; i++ {
		wg.Add(1)
		go func() {
			opt := filequeue.New("./config.json")
			for j := 0; j < 100; j++ {
				copy(add1.address[:], str[:])
				add1.txs = j * maxtestnum
				data := *((*([unsafe.Sizeof(add1)]byte))(unsafe.Pointer(&add1)))
				opt.D = data[:]
				for {
					err := opt.Send()
					if err != nil {
						continue
					} else {
						break
					}
				}
			}
			wg.Done()
		}()
	}

	//read
	wg.Add(1)
	go func() {
		opt := filequeue.New("./config.json")
		count := 0
		for {
			rd, _ := opt.Recv()
			if rd != nil {
				count++
				add2 = **(**addTxs)(unsafe.Pointer(&rd))
				if add2.address != add1.address {
					t.Fail()
				}
			}
			if rd == nil && count == maxtestnum {
				wg.Done()
				return
			}
		}

	}()
	wg.Wait()
}
