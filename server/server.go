package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-proxy/common"
	"go-proxy/common/network"

	"go-proxy/common/logs"
)

var (
	connectionPool     map[string]*ConnMatch
	connectionPoolLock sync.Mutex
	listenerPort       sync.Map
)

type ConnMatch struct {
	addTime time.Time
	accept  *net.TCPConn
	port    int64
}

func main() {

	var controlport int64
	var ServerHandlerKey string
	flag.Int64Var(&controlport, "p", 28009, "Server control port")

	flag.StringVar(&ServerHandlerKey, "k", "handler_key", "客户端认证key")
	flag.Parse()

	controlAddr := fmt.Sprintf("0.0.0.0:%d", controlport)
	logs.Infof("Listening control port: %s", controlAddr)

	connectionPool = make(map[string]*ConnMatch, 1024)
	logs.Info(common.GetCurrentDirectory())
	go createControlChannel(controlAddr, ServerHandlerKey)
	cleanConnectionPool()
}

// 创建一个控制通道，用于传递控制消息，如：心跳，创建新连接
func createControlChannel(controlAddr, serverhandlerkey string) {
	tcpListener, err := network.CreateTCPListener(controlAddr)
	if err != nil {
		logs.Fatal(err.Error())
		os.Exit(-1)
	}

	logs.Infof("Control listening: %s started successfully", controlAddr)

	for {
		var isauth bool
		t1 := time.Now()

		tcpConn, err := tcpListener.AcceptTCP()

		if err != nil {
			logs.Error(err.Error())
			continue
		}

		logs.Infof("NewConn: %s to control port", tcpConn.RemoteAddr().String())

		//10秒钟内未认证成功，则关闭连接
		go func() {
			for range time.Tick(10 * time.Second) {
				if time.Since(t1) > 10*time.Second {
					if !isauth {
						logs.Errorf("Get connection handler data error: %s ", tcpConn.RemoteAddr().String())
						tcpConn.Close()
						return
					}

				}
			}
		}()

		reader := bufio.NewReader(tcpConn)

		for {
			s, err := reader.ReadString('\n')
			if err != nil || err == io.EOF {
				logs.Errorf("Read remote control data error: %s", err.Error())
				time.Sleep(5 * time.Second)
				break
			}
			if !isauth {
				if s == serverhandlerkey+"\n" {
					isauth = true
					continue
				} else {
					logs.Errorf("Auth handler data error: %s", s)
					tcpConn.Close()
					break
				}

			}

			setport := strings.Split(strings.Replace(s, "\n", "", -1), ":")

			logs.Info("setport info ", setport)

			var port int64

			if len(setport) == 2 && setport[0] == "SETPORT" {
				port, err = strconv.ParseInt(setport[1], 10, 64)
				if err != nil {
					logs.Error("Set port error ", err.Error())
					setTunPortErr(tcpConn)
				}
				if port > 1000 {
					if !checkPortIsOpen(port) {
						sendMessage(network.SetTunnelERROR, tcpConn)
						tcpConn.Close()
						break
					} else {
						go keepAlive(tcpConn, port)
						break
					}

				}
			}
			tcpConn.Close()
			break

		}

	}
}

func keepAlive(Conn *net.TCPConn, port int64) {

	go AcceptUserRequest(port, Conn)
	go AcceptClientRequest(port)

	go func() {

		for range time.Tick(5 * time.Second) {
			if Conn == nil {
				logs.Error("Client Conn is nil:", Conn.RemoteAddr().String())
				closeListenerPort(port)
				return
			}
			_, err := Conn.Write(([]byte)(network.KeepAlive + "\n"))
			if err != nil {
				logs.Error("Client Conn stop:", Conn.RemoteAddr().String())
				closeListenerPort(port)
				Conn.Close()
				return
			}

		}
	}()
}

func checkPortIsOpen(port int64) bool {
	l, err := net.Listen("tcp", ":"+strconv.FormatInt(port, 10))
	if err != nil {
		logs.Errorf("Agent connect port is open, error: %s", err.Error())
		return false
	}
	defer l.Close()

	l, err = net.Listen("tcp", ":"+strconv.FormatInt(port+1, 10))
	if err != nil {
		logs.Errorf("User connect port is open, error: %s", err.Error())
		return false
	}
	defer l.Close()

	logs.Info("check port is open success:", port)

	return true
}

// 监听来自用户的请求
func AcceptUserRequest(port int64, controlConn *net.TCPConn) {

	visitaddr := "0.0.0.0:" + strconv.FormatInt(port+1, 10)

	tcpListener, err := network.CreateTCPListener(visitaddr)
	if err != nil {
		logs.Error("Create visit TCP listener error:", err.Error())
		return
	}
	defer tcpListener.Close()
	listenerPort.Store(port+1, tcpListener)

	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			logs.Error("AcceptUserRequest listener err:", err.Error())
			break
		}

		addConn2Pool(tcpConn, port)
		sendMessage(network.NewConnection, controlConn)
	}

}

// 接收客户端来的请求并建立隧道
func AcceptClientRequest(port int64) error {

	tunneladdr := "0.0.0.0:" + strconv.FormatInt(port, 10)
	tcpListener, err := network.CreateTCPListener(tunneladdr)
	if err != nil {
		logs.Error("acceptClientRequest err", err.Error())
		return err
	}
	defer tcpListener.Close()
	listenerPort.Store(port, tcpListener)

	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			logs.Error("AcceptClientRequest listener err:", err.Error())
			break
		}
		go establishTunnel(tcpConn, port)
	}
	return err
}

//客户端退出，关闭端口监听
func closeListenerPort(port int64) {
	logs.Infof("Listening port close: %d,%d", port, port+1)
	if v, ok := listenerPort.Load(port); ok {
		err := v.(*net.TCPListener).Close()
		if err != nil {
			logs.Error("close port err:", err.Error)
		}
	}
	if v, ok := listenerPort.Load(port + 1); ok {
		err := v.(*net.TCPListener).Close()
		if err != nil {
			logs.Error("close port err:", err.Error)
		}
	}
}

// 将用户来的连接放入连接池中
func addConn2Pool(accept *net.TCPConn, port int64) {
	connectionPoolLock.Lock()
	defer connectionPoolLock.Unlock()
	now := time.Now()

	connectionPool[strconv.FormatInt(now.UnixNano(), 10)] = &ConnMatch{now, accept, port}
}

// 发送给客户端新消息
func sendMessage(message string, controlConn *net.TCPConn) {
	// logs.Info("sendMessage:", message, controlConn.RemoteAddr().String())
	if controlConn == nil {
		logs.Info("No client connection")
		return
	}
	_, err := controlConn.Write([]byte(message + "\n"))
	if err != nil {
		logs.Error("send message error:", message)
	}
}

func establishTunnel(tunnel *net.TCPConn, port int64) {

	connectionPoolLock.Lock()

	defer connectionPoolLock.Unlock()

	for key, connMatch := range connectionPool {
		if connMatch.accept != nil {
			if connMatch.port == port {
				go network.Join2Conn(connMatch.accept, tunnel)
				delete(connectionPool, key)
				return

			}
		}
	}
	_ = tunnel.Close()
}

func cleanConnectionPool() {
	for {
		connectionPoolLock.Lock()
		for key, connMatch := range connectionPool {
			if time.Since(connMatch.addTime) > time.Second*10 {
				_ = connMatch.accept.Close()
				delete(connectionPool, key)
			}
		}
		connectionPoolLock.Unlock()
		time.Sleep(5 * time.Second)
	}
}

func setTunPortErr(Conn *net.TCPConn) (err error) {
	_, err = Conn.Write(([]byte)(network.SetTunnelERROR + "\n"))
	Conn.Close()
	return
}
