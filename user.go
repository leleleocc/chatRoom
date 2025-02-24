package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	user := &User{
		Name: conn.RemoteAddr().String(),
		Addr: conn.RemoteAddr().String(),
		C:    make(chan string),
		conn: conn,

		server: server,
	}
	//监听当前user的channelm，已有消息就发给客户端
	go user.ListenMessage()
	return user
}
func (this *User) Online() {
	//用户加入表中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()
	//若用户上线，将上线信息加到服务器的channel中
	this.server.ServerListenMessage(this, "online!")
}
func (this *User) Offline() {
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()
	//若用户上线，将上线信息加到服务器的channel中
	this.server.ServerListenMessage(this, "offline!!")
}

func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg + "\r\n"))
}

// 处理用户信息
func (this *User) DoMessage(msg string) {
	fmt.Println(msg)
	if msg == "who" { //私发该用户
		this.server.mapLock.Lock()
		for name, cli := range this.server.OnlineMap {
			onlineMsg := "[" + cli.Addr + "]" + cli.Name + ":online...\r\n"
			fmt.Println(name, cli)
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" { //重命名用户
		newName := strings.Split(msg, "|")[1]
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMsg("this name is already exist!\r\n")
		} else {

			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()
			this.Name = newName
			this.SendMsg("rename success:" + this.Name + "\r\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" { //私发消息
		toName := strings.Split(msg, "|")[1]
		if toName == "" {
			this.SendMsg("message format error!\r\n")
			return
		}
		toUser, ok := this.server.OnlineMap[toName]
		fmt.Println(toName, toUser)
		if !ok {
			this.SendMsg("user is not exist!\r\n")
			return
		}
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg("none message !\r\n")
			return
		}
		toUser.SendMsg(this.Name + "say:" + content + "\r\n")
	} else { //将用户信息加入到服务器的channel中
		this.server.ServerListenMessage(this, msg)
	}
}

func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\r\n"))
	}
}
