package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/pibigstar/go-proxy/common/logs"
)

var (
	localPort  int
	remotePort int
)

func init() {
	flag.IntVar(&localPort, "l", 5200, "the user link port")
	flag.IntVar(&remotePort, "r", 3333, "client listen port")
}

type client struct {
	conn net.Conn
	// 数据传输通道
	read  chan []byte
	write chan []byte
	// 异常退出通道
	exit chan error
	// 重连通道
	reConn chan bool
}

// 从Client端读取数据
func (c *client) Read(ctx context.Context) {
	// 如果10秒钟内没有消息传输，则Read函数会返回一个timeout的错误
	_ = c.conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	for {
		select {
		case <-ctx.Done():
			return
		default:
			data := make([]byte, 10240)
			n, err := c.conn.Read(data)
			if err != nil && err != io.EOF {
				if strings.Contains(err.Error(), "timeout") {
					// 设置读取时间为3秒，3秒后若读取不到, 则err会抛出timeout,然后发送心跳
					_ = c.conn.SetReadDeadline(time.Now().Add(time.Second * 3))
					c.conn.Write([]byte("pi"))
					continue
				}
				logs.Info("Client data read err")
				c.exit <- err
				return
			}

			// 收到心跳包,则跳过
			if data[0] == 'p' && data[1] == 'i' {
				logs.Info("Receive heartbeat packet")
				continue
			}
			c.read <- data[:n]
		}
	}
}

// 将数据写入到Client端
func (c *client) Write(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-c.write:
			_, err := c.conn.Write(data)
			if err != nil && err != io.EOF {
				c.exit <- err
				return
			}
		}
	}
}

type user struct {
	conn net.Conn
	// 数据传输通道
	read  chan []byte
	write chan []byte
	// 异常退出通道
	exit chan error
}

// 从User端读取数据
func (u *user) Read(ctx context.Context) {
	_ = u.conn.SetReadDeadline(time.Now().Add(time.Second * 200))
	for {
		select {
		case <-ctx.Done():
			return
		default:
			data := make([]byte, 10240)
			n, err := u.conn.Read(data)
			if err != nil && err != io.EOF {
				u.exit <- err
				return
			}
			u.read <- data[:n]
		}
	}
}

// 将数据写给User端
func (u *user) Write(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-u.write:
			_, err := u.conn.Write(data)
			if err != nil && err != io.EOF {
				u.exit <- err
				return
			}
		}
	}
}

func main() {
	flag.Parse()

	defer func() {
		err := recover()
		if err != nil {
			logs.Error(err)
		}
	}()

	clientListener, err := net.Listen("tcp", fmt.Sprintf(":%d", remotePort))
	if err != nil {
		panic(err)
	}
	logs.Info("Listening: %d Port, waiting client connection... \n", remotePort)
	// 监听User来连接
	userListener, err := net.Listen("tcp", fmt.Sprintf(":%d", localPort))
	if err != nil {
		panic(err)
	}
	logs.Info("Listening: %d Port, waiting user connection... \n", localPort)
	for {
		// 有Client来连接了
		clientConn, err := clientListener.Accept()
		if err != nil {
			panic(err)
		}
		logs.Info("Client: %s connection successfully", clientConn.RemoteAddr())

		client := &client{
			conn:   clientConn,
			read:   make(chan []byte),
			write:  make(chan []byte),
			exit:   make(chan error),
			reConn: make(chan bool),
		}

		userConnChan := make(chan net.Conn)
		go AcceptUserConn(userListener, userConnChan)

		go HandleClient(client, userConnChan)

		<-client.reConn

		logs.Info("wait to client reconnect")

	}
}

func HandleClient(client *client, userConnChan chan net.Conn) {
	ctx, cancel := context.WithCancel(context.Background())

	go client.Read(ctx)
	go client.Write(ctx)

	user := &user{
		read:  make(chan []byte),
		write: make(chan []byte),
		exit:  make(chan error),
	}

	defer func() {
		_ = client.conn.Close()
		_ = user.conn.Close()
		client.reConn <- true
	}()

	for {
		select {
		case userConn := <-userConnChan:
			user.conn = userConn
			go handle(ctx, client, user)
		case err := <-client.exit:
			logs.Error("Client connection error, close connection", err.Error())
			cancel()
			return
		case err := <-user.exit:
			// fmt.Println("user出现错误，关闭连接", err.Error())
			logs.Error("User connection error, close connection", err.Error())
			cancel()
			return
		}
	}
}

// 将两个Socket通道链接
// 1. 将从user收到的信息发给client
// 2. 将从client收到信息发给user
func handle(ctx context.Context, client *client, user *user) {
	go user.Read(ctx)
	go user.Write(ctx)

	for {
		select {
		case userRecv := <-user.read:
			// 收到从user发来的信息
			client.write <- userRecv
		case clientRecv := <-client.read:
			// 收到从client发来的信息
			user.write <- clientRecv

		case <-ctx.Done():
			return
		}
	}
}

// 等待user连接
func AcceptUserConn(userListener net.Listener, connChan chan net.Conn) {
	userConn, err := userListener.Accept()
	if err != nil {
		panic(err)
	}
	logs.Info("User: %s connection successfully", userConn.RemoteAddr())

	connChan <- userConn
}
