// Package handle
//
//	@Description:	响应设计方案：不使用http码来表示业务状态，采用业务状态码(code : msg)方式
//
// 只要到达后端的请求，http状态码都返回200，业务状态码为0表示成功，非0表示失败
// 只有后端发生panic，且被gin中间件捕获时，http状态码才返回500
package handle

import (
	"errors"
	"gin-blog-server/internal/global"
	"gin-blog-server/internal/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"log/slog"
	"net/http"
)

// Response
// Response[T any] 响应数据类型
//
//	@Description:	业务状态响应体
type Response[T any] struct {
	Code    int    `json:"code"`    //业务状态码
	Message string `json:"message"` //响应消息
	Data    T      `json:"data"`    //响应数据
}

// PageQuery
//
//	@Description:	分页获取数据
type PageQuery struct {
	Page    int    `form:"page_num"`  //页码(从1开始)
	Size    int    `form:"page_size"` //每页数量
	Keyword string `form:"keyword"`   //搜索关键字
}

// PageResult
// PageResult[T any]
//
//	@Description:分页响应数据
type PageResult[T any] struct {
	Page  int `json:"page_num"`  //页码(从1开始)
	Size  int `json:"page_size"` //每页数量
	Total int `json:"total"`     //总数
	List  []T `json:"page_data"` //分页数据
}

// ReturnHttpResponse
//
//	@Description:	返回http码、业务码、消息、数据
//	@Param			c			query	string	true	"上下文"
//	@Param			httpCode	query	int		true	"http码"
//	@Param			code		query	int		true	"业务码"
//	@Param			msg			query	string	true	"消息"
//	@Param			data		query	any		true	"数据"
func ReturnHttpResponse(c *gin.Context, httpCode, code int, msg string, data any) {
	c.JSON(httpCode, Response[any]{
		Code:    code,
		Message: msg,
		Data:    data,
	})
}

// ReturnResponse
//
//	@Description:返回业务码、数据
//	@Param	c		query	string	true	"上下文"
//	@Param	code	query	int		true	"业务码"
//	@Param	data	query	any		true	"数据"
func ReturnResponse(c *gin.Context, r global.Result, data any) {
	ReturnHttpResponse(c, http.StatusOK, r.Code(), r.Msg(), data)
}

// ReturnSuccess
//
//	@Description:	返回成功业务码加数据
//	@Param			c		query	string	true	"上下文"
//	@Param			data	query	any		true	"数据"
func ReturnSuccess(c *gin.Context, data any) {
	ReturnResponse(c, global.OKResult, data)
}

// ReturnError
//
//	@Description:错误分为业务错误与系统错误，业务层面上处理，返回200状态码
//
// 对于不可预料到的错误，会触发panic，由gin中间件捕获，返回500状态码
// err是业务错误，data是错误数据，可以是error或string
//
//	@Param	c		query	string	true			"上下文"
//	@Param	r		query	query	global.Result	true	"业务码响应体"
//	@Param	data	query	query	any				true	"错误数据"
func ReturnError(c *gin.Context, r global.Result, data any) {
	slog.Info("[Func-ReturnError]" + r.Msg())
	var val string = r.Msg()
	//  利用类型断言判断data的类型，如果是error就取error的错误信息，如果是string就取string的值
	//	如果 data == nil，说明没有附加信息，就只使用 r.Msg() 的默认错误提示。
	if data != nil {
		switch v := data.(type) {
		case error:
			val = v.Error()
		case string:
			val = v
		}
		slog.Error(val)
	}
	// 返回错误
	c.AbortWithStatusJSON(http.StatusOK,
		Response[any]{
			Code:    r.Code(),
			Message: r.Msg(),
			Data:    val,
		},
	)
}

// GetDB 获取 *gorm.DB
func GetDB(c *gin.Context) *gorm.DB {
	return c.MustGet(global.CTX_DB).(*gorm.DB)
}

// GetRDB 获取 *redis.Client
func GetRDB(c *gin.Context) *redis.Client {
	return c.MustGet(global.CTX_RDB).(*redis.Client)
}

// CurrentUserAuth
/*
获取当前登录用户信息
1. 能从 gin Context 上获取到 user 对象, 说明本次请求链路中获取过了
2. 从 session 中获取到 uid
3. 根据 uid 获取用户信息, 并设置到 gin Context 上
*/
func CurrentUserAuth(c *gin.Context) (*model.UserAuth, error) {
	key := global.CTX_USER_AUTH
	//1
	if cache, exist := c.Get(key); exist && cache != nil {
		slog.Debug("[Func-CurrentUserAuth] get from cache: " + cache.(*model.UserAuth).Username)
		return cache.(*model.UserAuth), nil
	}
	//2
	session := sessions.Default(c)
	id := session.Get(key)
	if id == nil {
		return nil, errors.New("session 中没有 user_auth_id")
	}
	//3
	db := GetDB(c)
	user, err := model.GetUserAuthInfoById(db, id.(int))
	if err != nil {
		return nil, err
	}
	c.Set(key, user)
	return user, nil
}
