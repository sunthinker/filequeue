package filequeue

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"unsafe"
)

type Conf struct {
	DataSize int    `json:"datasize"` //数据结构大小
	MaxNum   int    `json:"maxnum"`   //最大数据量maxnum*datasize
	FilePath string `json:"filepath"` //文件名(包含路径)
}

type FileOpt struct {
	Data  []byte
	mutex sync.Mutex
}

type Header struct {
	WriteSeek int64 `json:"writeseek"` //写文件位置指针
	ReadSeek  int64 `json:"readseek"`  //读文件位置指针
	Empty     bool  `json:"empty"`     //文件空
	Full      bool  `json:"full"`      //文件满
}

const (
	ConfFilePath = "./config.json"
)

//配置
var config Conf

//文件头
var head Header

//文件锁
var filemutex sync.Mutex

//初始化读取配置，重载文件头(seek)
func init() {
	rbuff := make([]byte, 128)

	//读取配置文件
	file, err := os.OpenFile(ConfFilePath, os.O_RDONLY, 0666)
	//加载配置失败，直接退出
	if err != nil {
		file.Close()
		panic("Loading Config Failed...")
	}
	n, _ := file.Read(rbuff)
	_ = json.Unmarshal(rbuff[:n], &config)
	//如果文件路径为空，则判断配置错误，直接退出
	if config.FilePath == "" {
		file.Close()
		panic("Config File Error...")
	}
	file.Close()

	//重载头(seek)
	file, err = os.OpenFile(config.FilePath, os.O_RDONLY, 0666)
	//defer file.Close()
	//重置文件
	if err != nil {
		//打开失败则新建(覆盖)
		file, _ := os.OpenFile(config.FilePath, os.O_CREATE|os.O_TRUNC, 0666)
		file.Close()
		//注意seek要偏移文件头位置
		head.WriteSeek = (int64)(unsafe.Sizeof(head))
		head.ReadSeek = (int64)(unsafe.Sizeof(head))
		head.Empty = true
		head.Full = false
		//创建新文件头
		new(FileOpt).SaveHead()
	} else {
		//读固定长度
		buf := make([]byte, unsafe.Sizeof(head))
		file.Read(buf)
		//注意[]byte的底层结构（切片）
		head = **(**Header)(unsafe.Pointer(&buf))
		file.Close()
	}

}

//写文件队列
func (fop *FileOpt) Send() {
	//fop.mutex.Lock()
	//fop.mutex.Unlock()
	//打开文件(创建)
	if head.Full == false {
		file, _ := os.OpenFile(config.FilePath, os.O_WRONLY, 0666)
		defer file.Close()
		n, err := file.WriteAt(fop.Data[0:config.DataSize], head.WriteSeek)
		if n == config.DataSize && err == nil {
			head.Empty = false
			//write seek移位，注意归0处理
			head.WriteSeek += (int64)(config.DataSize)
			head.WriteSeek %= (int64)(config.MaxNum) * (int64)(config.DataSize)
			if head.WriteSeek == 0 {
				head.WriteSeek = (int64)(unsafe.Sizeof(head))
			}
			//读/写seek相等时，文件满
			if head.WriteSeek == head.ReadSeek {
				head.Full = true
			} else {
				head.Full = false
			}
			new(FileOpt).SaveHead()
		} else {
			fmt.Println("Write Data Failed...")
		}
	} else {
		fmt.Println("The File Queue is Full..")
	}
}

//读文件队列
func (fop *FileOpt) Recv() []byte {
	if head.Empty == false {
		file, _ := os.OpenFile(config.FilePath, os.O_RDONLY, 0666)
		defer file.Close()
		//读固定长度
		buf := make([]byte, config.DataSize)
		n, err := file.ReadAt(buf, head.ReadSeek)
		if n == config.DataSize && err == nil {
			head.Full = false
			//read seek移位，注册归0处理
			head.ReadSeek += (int64)(config.DataSize)
			head.ReadSeek %= (int64)(config.MaxNum) * (int64)(config.DataSize)
			if head.ReadSeek == 0 {
				head.ReadSeek = (int64)(unsafe.Sizeof(head))
			}
			//读/写seek相等，文件空
			if head.ReadSeek == head.WriteSeek {
				head.Empty = true
			} else {
				head.Empty = false
			}
			new(FileOpt).SaveHead()
			//fop.Data = buf
			return buf
		} else {
			fmt.Println("Read Data Failed...")
			return nil
		}
	} else {
		fmt.Println("The File Queue is Empty")
		return nil
	}

}

//更新文件头(读/写的seek位置)
func (fop *FileOpt) SaveHead() {
	//打开文件
	file, _ := os.OpenFile(config.FilePath, os.O_WRONLY, 0666)
	defer file.Close()
	//将结构体类型的头，转换为[]byte类型
	data := *((*([unsafe.Sizeof(head)]byte))(unsafe.Pointer(&head)))
	//写头，固定长度
	file.WriteAt(data[0:unsafe.Sizeof(head)], 0)
}
