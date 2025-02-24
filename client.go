package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	Conn       net.Conn
	flag       int //client模式
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}
	client.Conn = conn
	return client
}

func (client *Client) Menu() bool {
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出系统")
	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println("请输入合法的选项")
		return false
	}
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.Menu() != true {
		}
		switch client.flag {
		case 1:
			client.publicChat()
		case 2:
			client.privateChat()
		case 3:
			client.updateName()
		}
	}
}

func (client *Client) dealResponse() {
	io.Copy(os.Stdout, client.Conn) //直接拷贝到标准输出,永久阻塞监听
}

func (client *Client) checkUser() {
	sendMsg := "who\r\n"
	_, err := client.Conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("client.Conn.Write error:", err)
		return
	}
}
func (client *Client) publicChat() {
	var chatMsg string
	fmt.Println("请输入聊天内容,exit退出")
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\r\n"
			_, err := client.Conn.Write([]byte(sendMsg)) //发送消息到服务器
			if err != nil {
				fmt.Println("client.Conn.Write error:", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println("请输入聊天内容,exit退出")
		fmt.Scanln(&chatMsg)
	}
}
func (client *Client) privateChat() {
	var remoteName, chatMsg string
	client.checkUser()
	fmt.Println("请输入聊天对象,exit退出")
	fmt.Scanln(&remoteName)
	for remoteName != "exit" {
		fmt.Println("请输入聊天内容,exit退出")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit" {
			sendMsg := "to|" + remoteName + "|" + chatMsg + "\r\n"
			_, err := client.Conn.Write([]byte(sendMsg)) //发送消息到服务器
			if err != nil {
				fmt.Println("client.Conn.Write error:", err)
				break
			}
			chatMsg = ""
			fmt.Println("请输入聊天内容,exit退出")
			fmt.Scanln(&chatMsg)
		}
		remoteName = ""
		fmt.Println("请输入聊天对象,exit退出")
		fmt.Scanln(&remoteName)
	}
}

func (client *Client) updateName() {
	fmt.Println("请输入用户名:")
	fmt.Scanln(&client.Name)
	sendMsg := "rename|" + client.Name + "\r\n"
	_, err := client.Conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("client.Conn.Write error:", err)
		return
	}
	return
}

var serverIp string
var serverPort int

// ./client -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址(默认127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口(默认8888)")
}

func main() {
	//命令行解析
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("client connect failed")
		return
	}
	go client.dealResponse() //接收服务器回执消息
	fmt.Println("client connect success")
	client.Run()
}
