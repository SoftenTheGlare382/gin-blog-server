package global

import (
	"flag"
	"testing"
)

func TestConfig(t *testing.T) {
	//允许用户通过命令行指定配置文件路径（如 -c ./myconfig.yml），
	//若不指定则使用默认路径 ./config.yml。
	configPath := flag.String("c", "../../config.yml", "配置文件路径")
	flag.Parse()

	// 读取配置文件
	conf := ReadConfig(*configPath)
	if conf.Server.Mode != "debug" {
		t.Errorf("配置文件加载失败")
	}

}
