##  Description
- The filequeue use to implement queue operations in a file.
- Supports multi-threaded asynchronous read and write.(But the efficiency to be optimized)
- Loop write and read
## Introduction of the config file
- data_size : The fixed size of each message.(byte)
- max_numm : Maximum number of messages that can be written.
- file_path : The file used to store messages.
## How to use
```golang
package main
import (
    "fmt"
    "github.com/sunthinker/filequeue"
)
func main() {
    opt := filequeue.New("config_path")
    opt.D = ...
    ...
    opt.Send()
    ...
    recv = opt.Recv()
    ...
}
```
## File structrue
- Header (Located at the beginning of the file for recording read and write seek)
- message_1.message_2...message_max_num (Each message body)
- write seek (Record the offset position of the write)
- read seek (Record the offset position of the read)
　
