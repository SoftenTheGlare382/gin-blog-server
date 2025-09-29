package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"testing"
)

func TestLoggerandRecovery(t *testing.T) {
	r := gin.Default()

	// 注册 Logger 中间件，记录日志
	r.Use(Logger())

	// 注册 Recovery 中间件，捕获 panic 错误并返回 500 错误
	r.Use(Recovery(false))

	// 定义一个路由，模拟一个 panic 错误
	r.GET("/panic", func(c *gin.Context) {
		panic("Something went wrong!")
	})

	// 启动服务器
	r.Run(":8080")
}

// 测试 session
func SetSession(c *gin.Context) {
	// 获取当前的 session
	session := sessions.Default(c)

	// 设置 session 数据，键值对形式
	session.Set("user", "JohnDoe")
	session.Set("isLoggedIn", true)

	// 提交 session 以保存数据
	if err := session.Save(); err != nil {
		c.JSON(500, gin.H{"error": "Failed to save session"})
		return
	}

	c.JSON(200, gin.H{"message": "Session data stored"})
}
func GetSession(c *gin.Context) {
	// 获取当前的 session
	session := sessions.Default(c)

	// 获取存储在 session 中的数据
	user := session.Get("user")
	isLoggedIn := session.Get("isLoggedIn")

	// 判断是否存在
	if user == nil {
		c.JSON(200, gin.H{"message": "User not logged in"})
		return
	}

	c.JSON(200, gin.H{
		"message":    "User is logged in",
		"user":       user,
		"isLoggedIn": isLoggedIn,
	})
}
func ClearSession(c *gin.Context) {
	session := sessions.Default(c)

	// 删除指定的 session 数据
	session.Delete("user")
	session.Delete("isLoggedIn")

	// 提交 session 以清除数据
	if err := session.Save(); err != nil {
		c.JSON(500, gin.H{"error": "Failed to clear session"})
		return
	}

	c.JSON(200, gin.H{"message": "Session data cleared"})
}
func TestWithCookieStore(t *testing.T) {
	r := gin.Default()
	r.Use(WithCookieStore("my_session", "secret"))
	r.GET("/set_session", SetSession)
	r.GET("/get_session", GetSession)
	r.GET("/clear_session", ClearSession)
	r.Run(":8080")
}
