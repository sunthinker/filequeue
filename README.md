# filequeue
Implement queue operations in a file（在文件中实现队列操作）
# config.json
datasize -> unsafe.Sizeof(var),变量占用的内存空间大小
# quick start
1、配置好config.json
2、
package main
import (
    "fmt"
    "github.com/sunthinker/filequeue"
)

func main() {
    var opt filequeue.FileOpt

    opt.Data = ...  //send data([]byte)
    opt.Send()
    recv = opt.Recv()
}
