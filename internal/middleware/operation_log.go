package middleware

import (
	"bytes"
	"gin-blog-server/internal/global"
	"gin-blog-server/internal/handle"
	"gin-blog-server/internal/model"
	"gin-blog-server/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"log/slog"
	"strings"
)

// TODO: 优化 API 路径格式
var optMap = map[string]string{
	"Article":      "文章",
	"BlogInfo":     "博客信息",
	"Category":     "分类",
	"Comment":      "评论",
	"FriendLink":   "友链",
	"Menu":         "菜单",
	"Message":      "留言",
	"OperationLog": "操作日志",
	"Resource":     "资源权限",
	"Role":         "角色",
	"Tag":          "标签",
	"User":         "用户",
	"Page":         "页面",
	"Login":        "登录",

	"POST":   "新增或修改",
	"PUT":    "修改",
	"DELETE": "删除",
}

func GetOptString(opt string) string {
	return optMap[opt]
}

// CustomResponseWriter 在 gin 中获取 Response Body 内容：对 gin 的 ResponseWriter 进行包装，每次响应数据时，将响应数据返回回去
type CustomResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer // 响应体缓存
}

// OperationLog 记录操作日志的中间件
func OperationLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 不记录 GET 请求操作记录 (太多了) 和 文件上传操作记录 (请求体太长)
		if c.Request.Method != "GET" && !strings.Contains(c.Request.RequestURI, "upload") {
			blw := &CustomResponseWriter{
				body:           bytes.NewBufferString(""),
				ResponseWriter: c.Writer,
			}
			c.Writer = blw

			auth, _ := handle.CurrentUserAuth(c)

			body, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

			ipAddress := "未知"
			ipSource := utils.IP.GetIpSource(ipAddress)

			moduleName := GetOptResource(c.HandlerName())
			operationLog := model.OperationLog{
				OptModule:     moduleName, // TODO : 优化
				OptType:       GetOptString(c.Request.Method),
				OptUrl:        c.Request.RequestURI,
				OptMethod:     c.HandlerName(),
				OptDesc:       GetOptString(c.Request.Method) + moduleName, // TODO: 优化
				RequestParam:  string(body),
				RequestMethod: c.Request.Method,
				UserId:        auth.UserInfoId,
				Nickname:      auth.UserInfo.Nickname,
				IpAddress:     ipAddress,
				IpSource:      ipSource,
			}
			c.Next()
			operationLog.ResponseData = blw.body.String() //从缓存中获取响应体内容

			db := c.MustGet(global.CTX_DB).(*gorm.DB)
			if err := db.Create(&operationLog).Error; err != nil {
				slog.Error("操作日志记录失败", err)
				handle.ReturnError(c, global.ErrDbOp, err)
				return
			}

		} else {
			c.Next()
		}
	}
}

// GetOptResource "gin-blog/api/v1.(*Resource).Delete-fm" => "Resource"
func GetOptResource(handleName string) string {
	s := strings.Split(handleName, ".")[1]
	return s[2 : len(s)-1]
}
