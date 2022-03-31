package filequeue

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"unsafe"
)

type Conf struct {
	DataSize int    `json:"data_size"` //the struct data size
	MaxNum   int    `json:"max_num"`   //DataSize*MaxNum
	FilePath string `json:"file_path"` //file path
}

type Header struct {
	WriteSeek int64 `json:"write_seek"` //write seek pos
	ReadSeek  int64 `json:"read_seek"`  //read seek pos
	Empty     bool  `json:"empty"`      //file is empty
	Full      bool  `json:"full"`       //file is full
}

type FileOpt struct {
	C Conf
	H Header
	D []byte
	L *sync.Mutex
}

var (
	optLock sync.Mutex
)

/*
 * new a Fileoption obj
 */
func New(conf string) *FileOpt {
	fopt := FileOpt{}
	fopt.L = &optLock

	fopt.L.Lock()
	rbuf := make([]byte, 128)
	//## 1、read config file
	f1, err := os.OpenFile(conf, os.O_RDONLY, 0666)
	defer f1.Close()
	//if open config file err then panic
	if err != nil {
		panic("Loading config failed.")
	}
	n, _ := f1.Read(rbuf)
	json.Unmarshal(rbuf[:n], &(fopt.C))
	//if parse file path err then panic
	if fopt.C.FilePath == "" {
		panic("File path is nil.")
	}
	//## 2、read file and reload head
	f2, err := os.OpenFile(fopt.C.FilePath, os.O_RDONLY, 0666)
	defer f2.Close()
	//if open file err, then reset file(os.CREATE|os.O_TRUNC)
	if err != nil {
		//create file
		f3, _ := os.OpenFile(fopt.C.FilePath, os.O_CREATE|os.O_TRUNC, 0666)
		defer f3.Close()
		//note: the init seek must offset head-size
		fopt.H.WriteSeek = (int64)(unsafe.Sizeof(fopt.H))
		fopt.H.ReadSeek = (int64)(unsafe.Sizeof(fopt.H))
		fopt.H.Empty = true
		fopt.H.Full = false
		//save init head
		fopt.SetHead()
	} else {
		//read
		buf := make([]byte, unsafe.Sizeof(fopt.H))
		f2.Read(buf)
		//[]byte convert to head struct
		fopt.H = **(**Header)(unsafe.Pointer(&buf))
	}
	fopt.L.Unlock()
	return &fopt
}

/*
 * write file queue
 */
func (fopt *FileOpt) Send() error {
	fopt.L.Lock()
	fopt.GetHead()
	if fopt.H.Full == false {
		//open file
		f, _ := os.OpenFile(fopt.C.FilePath, os.O_WRONLY, 0666)
		defer f.Close()
		n, err := f.WriteAt(fopt.D[0:fopt.C.DataSize], fopt.H.WriteSeek)
		//write ok then update head
		if n == fopt.C.DataSize && err == nil {
			fopt.H.Empty = false
			//cyclically move the write seek
			fopt.H.WriteSeek += (int64)(fopt.C.DataSize)
			fopt.H.WriteSeek %= (int64)(fopt.C.MaxNum) * (int64)(fopt.C.DataSize)
			if fopt.H.WriteSeek == 0 {
				fopt.H.WriteSeek = (int64)(unsafe.Sizeof(fopt.H))
			}
			//if read seek == write seek then file is full
			if fopt.H.WriteSeek == fopt.H.ReadSeek {
				fopt.H.Full = true
			} else {
				fopt.H.Full = false
			}
			fopt.SetHead()
		} else {
			//log.Println("Write Data Failed.")
			fopt.L.Unlock()
			return errors.New("Write Data Failed.")
		}
	} else {
		//log.Println("The File Queue is Full.")
		fopt.L.Unlock()
		return errors.New("The File Queue is Full.")
	}
	fopt.L.Unlock()
	return nil
}

/*
 * read file queue
 */
func (fopt *FileOpt) Recv() ([]byte, error) {
	fopt.L.Lock()
	fopt.GetHead()
	if fopt.H.Empty == false {
		f, _ := os.OpenFile(fopt.C.FilePath, os.O_RDONLY, 0666)
		defer f.Close()
		//read
		buf := make([]byte, fopt.C.DataSize)
		n, err := f.ReadAt(buf, fopt.H.ReadSeek)
		if n == fopt.C.DataSize && err == nil {
			fopt.H.Full = false
			//cyclically move the read seek
			fopt.H.ReadSeek += (int64)(fopt.C.DataSize)
			fopt.H.ReadSeek %= (int64)(fopt.C.MaxNum) * (int64)(fopt.C.DataSize)
			if fopt.H.ReadSeek == 0 {
				fopt.H.ReadSeek = (int64)(unsafe.Sizeof(fopt.H))
			}
			//if read seek == write seek then file is empty
			if fopt.H.ReadSeek == fopt.H.WriteSeek {
				fopt.H.Empty = true
			} else {
				fopt.H.Empty = false
			}
			fopt.SetHead()
			fopt.L.Unlock()
			return buf, nil
		} else {
			//log.Println("Read Data Failed.")
			fopt.L.Unlock()
			return nil, errors.New("Read Data Failed.")
		}
	} else {
		//log.Println("The File Queue is Empty")
		fopt.L.Unlock()
		return nil, errors.New("The File Queue is Empty.")
	}

}

/*
 * Update the head
 */
func (fopt *FileOpt) SetHead() {
	//open file
	f, _ := os.OpenFile(fopt.C.FilePath, os.O_WRONLY, 0666)
	defer f.Close()
	//convert struct to []byte
	data := *((*([unsafe.Sizeof(fopt.H)]byte))(unsafe.Pointer(&(fopt.H))))
	//write at pos 0(the head pos)
	f.WriteAt(data[0:unsafe.Sizeof(fopt.H)], 0)
}

/*
 * Get Head
 */
func (fopt *FileOpt) GetHead() {
	f, _ := os.OpenFile(fopt.C.FilePath, os.O_RDONLY, 0666)
	defer f.Close()
	//read
	buf := make([]byte, unsafe.Sizeof(fopt.H))
	f.Read(buf)
	//[]byte convert to head struct
	fopt.H = **(**Header)(unsafe.Pointer(&buf))
}
