package network

import (
	"bufio"
	"io"
	"net"
	"time"

	"go-proxy/common/logs"

	"go-proxy/common"
)

var (
	localServerAddr string

	remoteIP string
	// 远端的服务控制通道，用来传递控制信息，如出现新连接和心跳
	remoteControlAddr string
	// 远端服务端口，用来建立隧道
	remoteServerAddr string
)

func ClientRun() {
	localServerAddr = "127.0.0.1:" + common.Conf.Server.LocalPort

	remoteIP = common.Conf.Server.ServerHost

	// 远端的服务控制通道，用来传递控制信息，如出现新连接和心跳
	remoteControlAddr = remoteIP + ":" + common.Conf.Server.ServerControlPort
	// 远端服务端口，用来建立隧道
	remoteServerAddr = remoteIP + ":" + common.Conf.Server.ServerPort

	for {
		var checkstatus bool
		t1 := time.Now()

		tcpConn, err := CreateTCPConn(remoteControlAddr)

		if err != nil {
			logs.Errorf("Connection remote control server failed %s %s", remoteControlAddr, err.Error())
			time.Sleep(5 * time.Second)
			continue
		}

		logs.Infof("Connection server: %s success", remoteControlAddr)

		err = serverAuthHandler(tcpConn)

		if err != nil {
			logs.Error("server auth header failed", err.Error())
			time.Sleep(3 * time.Second)
			continue
		}
		logs.Infof("Auth handler: %s success", remoteControlAddr)

		err = setTunPort(tcpConn)

		if err != nil {
			logs.Error("Set tun port request failed", err.Error())
			time.Sleep(3 * time.Second)
			continue
		}

		reader := bufio.NewReader(tcpConn)

		for {
			if !checkstatus {
				checkstatus = true
				go func() {
					for range time.Tick(10 * time.Second) {
						if time.Since(t1) > 10*time.Second {
							logs.Info("keepalive timeout")
							tcpConn.Close()
							return
						}
					}
				}()
			}

			s, err := reader.ReadString('\n')
			if err != nil || err == io.EOF {
				logs.Errorf("Read remote control data error: %s", err.Error())
				time.Sleep(5 * time.Second)
				break
			}

			// logs.Infof(s)

			if s == SetTunnelERROR+"\n" {
				logs.Info("Set tunnel port error")
				time.Sleep(5 * time.Second)
				break
			}

			// 当有新连接信号出现时，新建一个tcp连接
			if s == NewConnection+"\n" {
				logs.Info("new connect LocalAndRemote")
				go connectLocalAndRemote()
			}
			if s == KeepAlive+"\n" {
				t1 = time.Now()
			}

		}
	}
}

func serverAuthHandler(Conn *net.TCPConn) (err error) {
	_, err = Conn.Write(([]byte)(AuthHandleData + "\n"))
	return
}

func setTunPort(Conn *net.TCPConn) (err error) {
	tundata := "SETPORT:" + common.Conf.Server.ServerPort
	_, err = Conn.Write(([]byte)(tundata + "\n"))
	return
}

func connectLocalAndRemote() {
	local := connectLocal()
	remote := connectRemote()

	if local != nil && remote != nil {
		Join2Conn(local, remote)
	} else {
		if local != nil {
			_ = local.Close()
		}
		if remote != nil {
			_ = remote.Close()
		}
	}
}

func connectLocal() *net.TCPConn {
	conn, err := CreateTCPConn(localServerAddr)
	if err != nil {
		logs.Infof("Connection to local port failed", err.Error())
	}
	return conn
}

func connectRemote() *net.TCPConn {
	conn, err := CreateTCPConn(remoteServerAddr)
	if err != nil {
		logs.Errorf("Connection to remote server: %s failed, error: %s ", remoteServerAddr, err.Error())
	}
	return conn
}
