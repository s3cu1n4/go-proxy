package main

import (
	"fmt"
	"go-proxy/common/network"
	"log"
	"os"

	"go-proxy/common"

	"github.com/kardianos/service"
)

var serviceConfig = &service.Config{
	Name:        "RemoteConnect",
	DisplayName: "Remote Connect Local Port Server",
	Description: "Connection remote server forward local remote desktop to remote",
}

func main() {

	common.GetConfig(common.GetCurrentDirectory() + "/conf.yaml")

	// 构建服务对象
	prog := &Program{}
	s, err := service.New(prog, serviceConfig)
	if err != nil {
		log.Fatal(err)
	}

	// 用于记录系统日志
	logger, err := s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) < 2 {
		err = s.Run()
		if err != nil {
			logger.Error(err)
		}
		return
	}

	cmd := os.Args[1]

	if cmd == "install" {
		err = s.Install()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("install successfully")
	}
	if cmd == "uninstall" {
		err = s.Uninstall()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("uninstall successfully")
	}

}

type Program struct{}

func (p *Program) Start(s service.Service) error {
	log.Println("start service")
	go p.run()
	return nil
}

func (p *Program) Stop(s service.Service) error {
	log.Println("stop service")
	return nil
}

func (p *Program) run() {
	network.ClientRun()
}
