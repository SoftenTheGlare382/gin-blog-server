package ginblog

import (
	"context"
	"gin-blog-server/internal/global"
	"gin-blog-server/internal/model"
	"github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"log/slog"
	"os"
	"time"
)

// InitLogger
//
//	@Description:	根据配置初始化日志
//	@Param			conf	body	global.Config	true	"配置对象"
//	@Return			*slog.Logger 日志对象

func InitLogger(conf *global.Config) *slog.Logger {
	//配置日志输出级别
	var level slog.Level
	switch conf.Log.Level {
	case "debug":
		level = slog.LevelDebug //debug 级别
	case "info":
		level = slog.LevelInfo //info 级别
	case "warn":
		level = slog.LevelWarn //warn 级别
	case "error":
		level = slog.LevelError //error 级别
	default:
		level = slog.LevelInfo //info 默认级别
	}
	//日志处理选项
	option := &slog.HandlerOptions{
		AddSource: false, //是否添加日志来源（例如，调用栈信息）
		Level:     level, //日志最小级别(debug | info | warn | error,默认info)
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			//定制化日志中的时间戳
			if a.Key == slog.TimeKey {
				//时间格式化为字符串，使用 Go 内建的时间格式
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format(time.DateTime))
				}
			}
			return a //返回处理后的属性
		},
	}
	// 日志格式选项
	var handler slog.Handler
	switch conf.Log.Format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, option) //使用JSON格式化日志
	case "text":
		fallthrough //使用文本格式化日志
	default:
		handler = slog.NewTextHandler(os.Stdout, option) //使用文本格式化日志
	}
	//创建新日志对象作为默认日志记录器并返回
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

// InitDatabase
//
//	@Description:初始化数据库连接，并返回一个GORM	DB实例，
//	根据配置文件中的设置来选择数据库类型，连接配置与日志级别，
//	需要时执行数据库迁移
//
//	@Param							conf	body	global.Config	true	"配置对象"
//	@Return							*gorm.DB GORM "DB实例"
func InitDatabase(conf *global.Config) *gorm.DB {
	dbtype := conf.DbType() //获取数据库类型( mysql | sqlite )
	dsn := conf.DbDSN()     //获取数据库连接字符串

	var db *gorm.DB
	var err error

	//配置数据库日志级别
	var level logger.LogLevel
	switch conf.Server.DbLogMode {
	case "silent":
		level = logger.Silent //静默模式
	case "warn":
		level = logger.Warn //警告模式
	case "info":
		level = logger.Info //信息模式
	case "error":
		fallthrough //错误模式
	default:
		level = logger.Error //默认错误模式
	}

	//配置GORM日志与其他参数
	config := &gorm.Config{
		Logger:                                   logger.Default.LogMode(level), //绑定到当前程序默认的日志系统，设置数据库日志级别
		DisableForeignKeyConstraintWhenMigrating: true,                          //禁用外键约束迁移
		SkipDefaultTransaction:                   true,                          //禁用默认事务
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, //禁用复数表名（例如：User 表名会映射到 `user`）
		},
	}
	//根据数据库类型选择不同数据库驱动进行数据库连接
	switch dbtype {
	case "mysql":
		//使用 MySQL 数据库
		db, err = gorm.Open(mysql.Open(dsn), config)
	case "sqlite":
		//使用 SQLite 数据库
		db, err = gorm.Open(sqlite.Open(dsn), config)
	default:
		//数据库类型不支持，退出程序
		log.Fatal("不支持的数据库类型: ", dbtype)
	}

	//检查数据库连接是否成功
	if err != nil {
		log.Fatal("数据库连接失败: ", err) //输出错误信息并退出程序
	}
	log.Println("数据库连接成功", dbtype, dsn)
	//是否开启数据库迁移
	if conf.Server.DbAutoMigrate {
		//运行数据库迁移
		if err := model.MakeMigrate(db); err != nil {
			log.Fatal("数据库迁移失败: ", err)
		}
		log.Println("数据库迁移成功")
	}
	return db
}

// InitRedis
//
//	@Description:	初始化redis客户端并测试连接
//	@Param			conf	body	global.Config	true	"配置对象"
//	@Return			*redis.Client "redis客户端对象实例"
func InitRedis(conf *global.Config) *redis.Client {
	//创建一个 Redis 客户端实例并配置参数
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Addr,     //Redis 服务器地址,形如：127.0.0.1:6379
		Password: conf.Redis.Password, //Redis 密码,如果没有设置密码则传空字符串
		DB:       conf.Redis.DB,       //Redis 数据库引索通常是 0、1、2 等，从配置文件中获取
	})
	//测试连接
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("Redis 连接失败: ", err)
	}
	log.Println("Redis 连接成功", conf.Redis.Addr, conf.Redis.DB, conf.Redis.Password)
	//返回 Redis 客户端对象实例
	return rdb
}
