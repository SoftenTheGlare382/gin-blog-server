package ginblog

import (
	"gin-blog-server/internal/global"
	"testing"
)

func TestHelper(t *testing.T) {
	global.ReadConfig("../config.yml")
	logger := InitLogger(global.GetConfig())
	logger.Info("gin-blog-server启动中...")
	logger.Debug("gin-blog-server启动中...")
	logger.Error("gin-blog-server启动中...")
	logger.Warn("gin-blog-server启动中...")
	if global.Conf.DbType() != "mysql" {
		t.Error("数据库类型错误")
	}
	if global.Conf.DbDSN() != "root:root@tcp(127.0.0.1:3306)/gin_blog_db?charset=utf8mb4&parseTime=True&loc=Local" {
		t.Error("数据库DSN错误")
	}
	InitRedis(global.GetConfig())
	InitDatabase(global.GetConfig())

}
