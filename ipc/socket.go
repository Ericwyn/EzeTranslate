package ipc

import (
	"fmt"
	"github.com/Ericwyn/EzeTranslate/log"
	"net"
	"os"
	"time"
)

type UnixSocket struct {
	filename string
	bufsize  int
	handler  func(string) string
}

func NewUnixSocket(filename string, size ...int) *UnixSocket {
	size1 := 10480
	if size != nil {
		size1 = size[0]
	}
	us := UnixSocket{filename: filename, bufsize: size1}
	return &us
}

func (this *UnixSocket) createServer() {
	os.Remove(this.filename)
	addr, err := net.ResolveUnixAddr("unix", this.filename)
	if err != nil {
		panic("Cannot resolve unix addr: " + err.Error())
	}
	listener, err := net.ListenUnix("unix", addr)
	defer listener.Close()
	if err != nil {
		panic("Cannot listen to unix domain socket: " + err.Error())
	}
	fmt.Println("Listening on", listener.Addr())
	for {
		c, err := listener.Accept()
		if err != nil {
			panic("Accept: " + err.Error())
		}
		go this.HandleServerConn(c)
	}

}

//接收连接并处理
func (this *UnixSocket) HandleServerConn(c net.Conn) {
	defer c.Close()
	buf := make([]byte, this.bufsize)
	nr, err := c.Read(buf)
	if err != nil {
		panic("Read: " + err.Error())
	}
	// 这里，你需要 parse buf 里的数据来决定返回什么给客户端
	// 假设 respnoseData 是你想返回的文件内容
	result := this.HandleServerContext(string(buf[0:nr]))
	_, err = c.Write([]byte(result))
	if err != nil {
		panic("Writes failed.")
	}
}

func (this *UnixSocket) SetContextHandler(f func(string) string) {
	this.handler = f
}

//接收内容并返回结果
func (this *UnixSocket) HandleServerContext(context string) string {
	if this.handler != nil {
		return this.handler(context)
	}
	now := time.Now().String()
	return now
}

func (this *UnixSocket) StartServer() {
	this.createServer()
}

//客户端
func (this *UnixSocket) ClientSendContext(context string) (string, error) {
	addr, err := net.ResolveUnixAddr("unix", this.filename)
	if err != nil {
		log.D("Cannot resolve unix addr: " + err.Error())
		return "", err
	}
	//拔号
	c, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		log.D("DialUnix failed.")
		return "", err
	}
	//写出
	_, err = c.Write([]byte(context))
	if err != nil {
		log.D("Writes failed.")
		return "", err
	}
	//读结果
	buf := make([]byte, this.bufsize)
	nr, err := c.Read(buf)
	if err != nil {
		log.D("Read: " + err.Error())
		return "", err
	}
	return string(buf[0:nr]), nil
}
