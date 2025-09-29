package global

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"strings"
)

// Config
//
//	@Description:应用程序所需的配置项，涵盖了多种模块（如服务器、数据库、日志、JWT、邮件、Redis、上传等）的配置信息。
type Config struct {
	//
	//  Server
	//	@Description:	服务器配置
	//
	Server struct {
		Mode          string // 服务器模式(debug | release)
		Port          string // 服务器端口号
		DbType        string // 数据库类型
		DbAutoMigrate bool   // 数据库表结构是否自动迁移
		DbLogMode     string // 数据库日志模式（silent | error | warn | info）
	}
	//
	//  Log
	//	@Description:	日志配置
	//
	Log struct {
		Level     string // 日志级别 (debug | info | warn | error )
		Prefix    string // 日志前缀
		Format    string // 日志格式 (json | text)
		Directory string // 日志目录
	}
	//
	//  JWT
	//	@Description:JWT配置
	//
	JWT struct {
		Secret string // JWT密钥
		Expire int64  // JWT过期时间(hours)
		Issuer string // JWT签发者
	}
	//
	//  Mysql
	//	@Description:Mysql数据库配置
	Mysql struct {
		Host     string //Mysql 服务器地址
		Port     string //Mysql 服务器端口
		Config   string //Mysql 高级配置
		Dbname   string //Mysql 数据库名称
		Username string //Mysql 用户名
		Password string //Mysql 密码
	}
	//
	//  SQLite
	//	@Description:SQLite数据库配置
	SQLite struct {
		Dsn string // SQLite 数据源名称（DSN）
	}
	//
	//  Redis
	//	@Description:Redis数据库配置
	Redis struct {
		DB       int    // Redis 数据库引索
		Addr     string // Redis 服务器地址:端口
		Password string // Redis 密码
	}
	//
	//  Session
	//	@Description:Session配置
	Session struct {
		Name   string //session 名称
		Salt   string //session 盐值
		MaxAge int    //session 过期时间（seconds）
	}
	//
	//  Email
	//	@Description:Email配置
	Email struct {
		From     string //发件人邮箱
		Host     string //SMTP 服务器地址
		Port     int    //SMTP 端口(默认 465)
		SmtpPass string //SMTP 密钥（开启SMTP时获取的密钥，而非邮箱密码）
		SmtpUser string //SMTP 用户名(邮箱账号)
	}
	//
	//  Captcha
	//	@Description:验证码配置
	Captcha struct {
		SendEmail  bool //是否发送邮件验证码
		ExpireTime int  //验证码过期时间(seconds)
	}
	//
	//  Upload
	//	@Description:文件上传配置
	Upload struct {
		Size      int    //上传文件大小限制(字节)
		OssType   string //OSS存储类型(local | qiniu)
		Path      string //本地文件访问路径
		StorePath string //本地文件存储路径
	}
	//
	//  Qiniu
	//	@Description:七牛云配置
	Qiniu struct {
		ImgPath       string //外链图片访问路径
		Zone          string //存储区域
		Bucket        string //空间名称
		AccessKey     string //七牛云访问密钥
		SecretKey     string //七牛云私密密钥
		UseHTTPS      bool   //是否使用https
		UseCdnDomains bool   //是否使用CDN上传加速
	}
}

// Conf 存储应用配置的全局变量
var Conf *Config

// GetConfig
//
//	@Description:返回全局配置，如果未初始化，则触发	panic
//	@return							*Config 全局配置
func GetConfig() *Config {
	if Conf == nil {
		log.Panic("配置文件未初始化")
		return nil
	}
	return Conf
}

// ReadConfig
//
//	@Description:	取并解析配置文件，并返回解析后的配置对象
//	@Param			path	path	string	true	"配置文件路径"
//	@Return			*Config  解析后的配置对象
func ReadConfig(path string) *Config {
	v := viper.New()
	v.SetConfigFile(path)                              //设置配置文件
	v.AutomaticEnv()                                   //从环境变量读取配置
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) //转换环境变量格式：SERVER_APP_MODE => SERVER.APPMODE
	//  读取配置文件
	if err := v.ReadInConfig(); err != nil {
		panic("读取配置文件失败：" + err.Error())
	}
	// 解析配置文件(绑定结构体,反序列化)
	if err := v.Unmarshal(&Conf); err != nil {
		panic("解析配置文件失败：" + err.Error())
	}
	log.Println("配置文件加载成功：", path)
	return Conf
}

// DbType
//
//	@Description:返回数据库类型，默认为	sqlite
//	@receiver					*Config 配置对象
//	@return						string 数据库类型
func (*Config) DbType() string {
	if Conf.Server.DbType == "" {
		Conf.Server.DbType = "sqlite"
	}
	return Conf.Server.DbType
}
func (*Config) DbDSN() string {
	switch Conf.Server.DbType {
	case "mysql":
		// 构造mysql连接字符串
		conf := Conf.Mysql
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
			conf.Username, conf.Password, conf.Host, conf.Port, conf.Dbname, conf.Config)
	case "sqlite":
		// 构造sqlite连接字符串
		conf := Conf.SQLite
		return conf.Dsn
	default:
		// 默认使用sqlite, 并使用内存数据库
		Conf.Server.DbType = "sqlite"
		if Conf.SQLite.Dsn == "" {
			Conf.SQLite.Dsn = "file::memory" //使用内存数据库
		}
		return Conf.SQLite.Dsn
	}
}
