package filequeue_test

import (
	"github.com/sunthinker/filequeue"
	"testing"
	"unsafe"
)

type addTxs struct {
	address [42]byte
	txs     int
}

var opt filequeue.FileOpt
var add1 addTxs
var add2 addTxs

const (
	str        = "0x81b7e08f65bdf5648606c89998a9cc8164397647"
	maxtestnum = 10000
)

func TestFileWR(t *testing.T) {
	for i := 0; i < maxtestnum; i++ {
		copy(add1.address[:], str[:])
		add1.txs = i
		//将结构体类型，转换为[]byte类型
		data := *((*([unsafe.Sizeof(add1)]byte))(unsafe.Pointer(&add1)))
		opt.Data = data[:]

		opt.Send()
		rd := opt.Recv()

		//将[]byte类型数据转为结构体类型
		add2 = **(**addTxs)(unsafe.Pointer(&rd))
		//比读写是否一致
		if add2.address != add1.address || add2.txs != add1.txs {
			t.Fail()
		}
	}
}
