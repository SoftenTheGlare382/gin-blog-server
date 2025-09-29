package main

import (
	"flag"
	_ "gin-blog-server/docs"
	ginblog "gin-blog-server/internal"
	"gin-blog-server/internal/global"
	"gin-blog-server/internal/middleware"
	"github.com/gin-gonic/gin"
	"log"
	"strings"
)

func main() {
	configPath := flag.String("c", "./config.yml", "配置文件路径")
	flag.Parse()

	// 根据文件路径读取配置文件
	conf := global.ReadConfig(*configPath)
	_ = ginblog.InitLogger(conf)
	db := ginblog.InitDatabase(conf)
	rdb := ginblog.InitRedis(conf)

	//初始化gin服务
	gin.SetMode(conf.Server.Mode)
	r := gin.New()
	r.SetTrustedProxies([]string{"*"}) //设置 Gin 框架信任所有的代理（Proxy）

	// 开发模式使用 gin 自带的日志和恢复中间件, 生产模式使用自定义的中间件
	if conf.Server.Mode == "debug" {
		r.Use(gin.Logger(), gin.Recovery())
	} else {
		r.Use(middleware.Logger(), middleware.Recovery(true))
	}

	r.Use(middleware.WithGormDB(db), middleware.WithRedisDB(rdb), middleware.CORS(), middleware.WithCookieStore(conf.Session.Name, conf.Session.Salt))
	ginblog.RegisterHandlers(r)

	//使用本地文件上传服务
	if conf.Upload.OssType == "local" {
		r.Static(conf.Upload.Path, conf.Upload.StorePath)
	}

	serverAddr := conf.Server.Port
	if serverAddr[0] == ':' || strings.HasPrefix(serverAddr, "0.0.0.0:") {
		log.Printf("Serving HTTP on (http://localhost:%s/) ... \n", strings.Split(serverAddr, ":")[1])
	} else {
		log.Printf("Serving HTTP on (http://%s/) ... \n", serverAddr)
	}

	r.Run(serverAddr)
}
