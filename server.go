package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int
	//在线用户
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	//消息广播
	Message chan string
}

func NewServer(ip string, port int) *Server {
	return &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
}
func (this *Server) BroadCast() string {
	for {
		msg := <-this.Message
		this.mapLock.RLock()

		for _, client := range this.OnlineMap {
			client.C <- msg
		}
		this.mapLock.RUnlock()
	}
}

func (this *Server) Handler(conn net.Conn) {
	//fmt.Println("connect success")
	user := NewUser(conn, this)
	//用户上线
	user.Online()
	//监听用户消息
	isLive := make(chan bool)
	go func() {
		reader := bufio.NewReader(conn)
		for {
			msg, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					// 连接关闭
					fmt.Println("Connection closed by client:", conn.RemoteAddr())
				} else {
					// 其他读取错误
					fmt.Println("conn.Read error:", err)
				}
				return // 错误发生时退出 goroutine
			}
			msg = strings.TrimSpace(msg)
			if msg == "" {
				continue
			}
			user.DoMessage(msg)
			isLive <- true
		}
	}()
	//阻塞handler
	for {
		select { //监听两个channel
		case <-isLive:

		case <-time.After(60 * 5 * time.Second): //After是chan,超时就有数据了
			user.SendMsg("you were forced to offline!!!")
			close(user.C)
			conn.Close()
			return //runtime.Goexit()
		}
	}
}

func (this *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen error:", err)
		return
	}
	defer listener.Close()
	//启动监听Message的goroutine,将C中消息广播给user的channel中
	go this.BroadCast()
	//死循环，等待客户端连接，连接成功就开辟一个goroutine处理业务
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept error:", err)
			continue
		}
		go this.Handler(conn) // 处理业务
	}
}

func (this *Server) ServerListenMessage(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ": " + msg
	this.Message <- sendMsg
}
