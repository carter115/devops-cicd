package main

import (
	"lyyops-cicd/config"
	"lyyops-cicd/dao"
	"lyyops-cicd/handler"
	"lyyops-cicd/pkg/log"
	"net/http"
)

// @title LYY CICD API
// @version 0.1
// @description 应用自动化部署
func main() {
	// 加载配置
	if err := config.InitConfig("config/config.yaml"); err != nil {
		panic(err)
	}

	// 初始化日志
	log.NewLogger(config.Config.Log)
	log.Infof("load config: %+v", config.Config)

	// 连接Redis
	if err := dao.InitRedis(); err != nil {
		log.Fatalf("init redis connection error: %+v", err)
	}

	// http server
	engine := handler.InitHandler()
	server := &http.Server{
		Addr:         config.Config.Server.Address,
		ReadTimeout:  config.Config.Server.ReadTimeout,
		WriteTimeout: config.Config.Server.WriteTimeout,
		Handler:      engine,
	}
	// 启动http server
	log.Info("http server is running")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("http server is closed: %v", err)
	}

}
