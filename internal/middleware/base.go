package middleware

import (
	"errors"
	"gin-blog-server/internal/global"
	"gin-blog-server/internal/handle"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

// WithGormDB
//
//	@Summary		将 GORM 数据库连接注入 Gin 上下文
//	@Description	中间件工厂函数，用于将 *gorm.DB 实例注入到 Gin 请求上下文中，以便后续处理函数能够访问数据库连接。
//
//	@Return			gin.HandlerFunc 返回一个 Gin 中间件函数，可用于注册到路由组中
func WithGormDB(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(global.CTX_DB, db)
		ctx.Next()
	}
}

// WithRedisDB
//
//	@Summary		将 Redis 客户端连接注入 Gin 上下文
//	@Description	中间件工厂函数，用于将 *redis.Client 实例注入到 Gin 请求上下文中，以便后续处理函数能够访问 Redis 连接。
//
//	@Return			gin.HandlerFunc 中间件函数，用于注册到 Gin 路由
func WithRedisDB(rdb *redis.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(global.CTX_RDB, rdb)
		ctx.Next()
	}
}

// WithCookieStore
//
//	@Summary		基于 Cookie 的会话管理中间件
//	@Description	创建基于 Cookie 的会话存储中间件，并将会话数据加密存储在客户端 Cookie 中。用于 Gin 路由中实现会话管理。
//	@Param			name	path	string	true	"会话名称，用于标识和存储会话"
//	@Param			secret	query	string	true	"加密密钥，用于安全签名 Cookie"
//	@Return			gin.HandlerFunc 中间件函数，用于注册到 Gin 路
func WithCookieStore(name, secret string) gin.HandlerFunc {
	//创建一个新的Cookie存储实例，使用指定的密钥（secret）进行加密
	store := cookie.NewStore([]byte(secret))
	//设置Cookie存储的选项，如路径、最大年龄等
	store.Options(sessions.Options{
		Path:   "/", // 指定 Cookie 的有效路径，"/" 表示整个网站都有效
		MaxAge: 600, // Cookie 的最大年龄，600秒，即10分钟
	})
	//  返回一个 Gin 中间件，使用 sessions.Sessions 方法将会话存储配置应用到 Gin 路由中。
	//`name` 是会话的名称，`store` 是用于存储会话数据的 Cookie 存储实例。

	return sessions.Sessions(name, store)
}

// CORS
//
//	@Summary		跨域请求处理中间件
//	@Description	创建并返回一个 CORS 中间件，用于处理跨域请求。支持自定义源、方法、请求头等配置，适用于 Gin 框架路由。
//	@Return			gin.HandlerFunc 返回一个 Gin 的中间件函数，用于注册到路由中
func CORS() gin.HandlerFunc {
	//  cors.New()创建并返回一个跨域处理中间件
	return cors.New(cors.Config{
		// 配置允许的源，此时为允许所有源
		AllowOrigins: []string{"*"},

		// 配置了允许跨域请求使用的HTTP方法。这里设置了常见的HTTP方法：PUT、POST、GET、DELETE、OPTIONS 和 PATCH
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},

		//  AllowHeaders 配置了允许的请求头，表示在跨域请求中可以携带哪些请求头。
		AllowHeaders: []string{"Origin", "X-Requested-With", "Content-Type", "Authorization"},

		// 配置了允许客户端访问的响应头，表示在跨域请求中可以暴露哪些响应头。
		ExposeHeaders: []string{"Content-Type"},

		// 是否允许跨域请求时携带用户凭证如Cookie
		AllowCredentials: true,

		// AllowOriginFunc 是一个函数，允许你对源进行动态验证，决定是否允许该源进行跨域请求。此处设置为总是返回true，表示任何源都被允许。
		AllowOriginFunc: func(origin string) bool {
			return true
		},

		// MaxAge 设置了预检请求的缓存时间，单位为时间.Duration。这里设置为24小时，表示预检请求结果会被缓存24小时。
		MaxAge: 12 * time.Hour,
	})
}

// Logger
// 日志记录
//
//	@Description:中间件函数，记录日志信息
//	@Return	gin.HandlerFunc 返回一个 gin.HandlerFunc 类型的中间件函数
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		start := time.Now()

		// 请求之前
		c.Next()
		// 请求之后

		// 计算请求处理耗时
		cost := time.Since(start)
		//  slog记录日志信息
		slog.Info("[GIN]", //日志标识，方便区分是来自 Gin 中间件的日志
			slog.String("path", c.Request.URL.Path),      //请求路径
			slog.String("query", c.Request.URL.RawQuery), //请求查询参数
			slog.Int("status", c.Writer.Status()),        //响应状态码
			slog.String("method", c.Request.Method),      //请求方法
			slog.String("ip", c.ClientIP()),              //客户端 IP
			slog.Int("size", c.Writer.Size()),            //响应大小
			slog.Duration("cost", cost),                  //请求处理耗时
			//需要时开启
			//slog.String("user-agent", c.Request.UserAgent()),                      //请求的用户代理
			//slog.String("body", c.Request.PostForm.Encode()),                      //请求体内容
			//slog.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()), //错误信息
		)

	}
}

// Recovery
// 错误恢复
//
//	@Description:	用于捕获处理程序中的panic信息
//	@Return			stack 参数控制是否打印panic的堆栈信息
func Recovery(stack bool) gin.HandlerFunc {
	//返回一个gin.HandlerFunc类型函数
	return func(c *gin.Context) {
		//使用defer语句，确保函数执行结束时会调用panic处理函数
		defer func() {
			//使用recover()函数，捕获panic信息，如果有panic信息，err将不为nil
			if err := recover(); err != nil {
				//检查是否网络连接错误（例如：Broken Pipe），这些错误不需要堆栈信息
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					var se *os.SyscallError
					//判断是否是网络连接错误
					if errors.As(ne, &se) {
						seStr := strings.ToLower(se.Error())
						if strings.Contains(seStr, "broken pipe") ||
							strings.Contains(seStr, "connection reset by peer") {
							brokenPipe = true
						}
					}
				}
				//如果发生panic，返回给客户端通用错误信息
				handle.ReturnHttpResponse(c, http.StatusInternalServerError, global.FAIL, global.GetMsg(global.FAIL), err)
				//打印发生panic时的http请求信息
				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				//如果是网络连接错误，则不返回堆栈信息,直接记录错误并终止
				if brokenPipe {
					//记录错误日志
					slog.Error(c.Request.URL.Path,
						slog.Any("error", err),
						slog.String("request",
							string(httpRequest)))
					//如果连接已经断开，不能再写入响应状态
					_ = c.Error(err.(error))
					c.Abort()
					return
				}
				//				//如果是其他错误，则返回堆栈信息
				if stack {
					slog.Error("[Recovery from panic]",
						slog.Any("error", err),
						slog.String("request", string(httpRequest)),
						slog.String("stack", string(debug.Stack())))
				} else {
					//					//如果不需要打印堆栈信息，则记录错误日志
					slog.Error("[Recovery from panic]",
						slog.Any("error", err),
						slog.String("request", string(httpRequest)),
					)
				}
				//保证请求被立即终止，只返回状态码
				c.AbortWithStatus(http.StatusInternalServerError)
			}

		}()
		c.Next()
	}
}
