/*

LICENSE:  MIT
Author:   sine
Email:    sinerwr@gmail.com

*/

package controller

import (
	"github.com/getsentry/raven-go"
	"google.golang.org/grpc"
	"log"

	"github.com/SiCo-Ops/Pb"
	"github.com/SiCo-Ops/cfg"
	"github.com/SiCo-Ops/dao/mongo"
)

const (
	configPath string = "config.json"
)

var (
	config            cfg.ConfigItems
	RPCServer         = grpc.NewServer()
	hookDB, hookDBErr = mongo.NewDial()
)

func ServePort() string {
	return config.RpcCPort
}

func init() {
	data, err := cfg.ReadFilePath(configPath)
	if err != nil {
		data = cfg.ReadConfigServer()
		if data == nil {
			log.Fatalln("config.json not exist and configserver was down")
		}
	}
	cfg.Unmarshal(data, &config)

	hookDB, hookDBErr = mongo.InitDial(config.MongoHookAddress, config.MongoHookUsername, config.MongoHookPassword)
	if hookDBErr != nil {
		log.Fatalln(hookDBErr)
	}
	err = mongo.HookEnsureIndexes(hookDB)
	if err != nil {
		log.Fatalln(err)
	}

	pb.RegisterHookServiceServer(RPCServer, &HookService{})

	if config.SentryCStatus == "active" && config.SentryCDSN != "" {
		raven.SetDSN(config.SentryCDSN)
	}
}
