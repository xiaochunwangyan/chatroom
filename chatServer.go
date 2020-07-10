package main

import (
	"fmt"
	"net"
	"strings"
)

//处理错误func
func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

//定义用户类
type Client struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn
}

//定义map用来存储用户信息
var onlineMap = make(map[string]Client)

var message = make(chan string)

var stop = make(chan int)

func main() {
	//开启监听
	listen, err := net.Listen("tcp", "192.168.60.216:8000")
	defer listen.Close()
	checkError(err)
	//新开一个协程 转发消息 对map进行遍历
	go TransferMessage()
	for {
		//处理用户连接
		conn, err := listen.Accept()
		checkError(err)
		//新开一个协程 将用户加入到map
		go HandleConn(conn)
	}

}

//广播在线信息
func TransferMessage() {
	for {
		//没有消息会进行阻塞
		mess := <-message
		for _, cli := range onlineMap {
			cli.C <- mess
		}
	}
}

//处理用户链接
func HandleConn(conn net.Conn) {
	defer conn.Close()
	//获取客户端的网络地址
	addr := conn.RemoteAddr().String()
	cli := Client{addr, addr, make(chan string), conn}
	onlineMap[addr] = cli
	//新开一个协程 给当前客户端发送消息
	go WriteToClient(cli, conn)
	//广播在线
	message <- MakeMess(cli, "login")
	//群发消息
	go func() {
		buf := make([]byte, 2048)
		for {
			read, err := conn.Read(buf)
			checkError(err)
			if read == 0 {
				fmt.Println("客户端退出")
				return
			}
			//获取到客户端输入的内容
			nameString := string(buf[:read])
			fmt.Println(nameString[:1])
			//显示所有在线用户
			if len(nameString) > 8 && nameString[:8] == "all_user" {
				for _, user := range onlineMap {
					user_info := user.Name + "\n"
					conn.Write([]byte(user_info))
				}
				//修改昵称
			} else if len(nameString) > 6 && nameString[:6] == "rename" {
				realName := nameString[7:]
				cli.Name = realName
				onlineMap[addr] = cli
				//指定特定的人进行发送消息
			} else if nameString[:1] == "@" {
				realName := nameString[1 : len(nameString)-1]
				fmt.Println("realName:" + realName)
				for _, user := range onlineMap {
					//判断某个@后面的名字是否是map里面的其中一个
					real := user.Name[:len(user.Name)-1]
					fmt.Println("real:" + real)
					if strings.Contains(realName, real) {
						fmt.Println("我进来了")
						c := onlineMap[user.Addr].conn
						//获取到发送的内容
						s := realName[len(user.Name):] + "\n"
						c.Write([]byte(s))
					}
				}
			} else {
				message <- MakeMess(cli, nameString)
			}
		}
	}()
	//防止退出
	<-stop
}

func MakeMess(cli Client, msg string) (buf string) {
	return "[" + cli.Addr + "]" + cli.Name + ":" + msg
}

func WriteToClient(cli Client, conn net.Conn) {
	for msg := range cli.C {
		conn.Write([]byte(msg + "\n"))
	}
}
