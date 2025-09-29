package ginblog

import (
	"gin-blog-server/docs"
	"gin-blog-server/internal/handle"
	"gin-blog-server/internal/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var (
	userAuthAPI     handle.UserAuth
	blogInfoAPI     handle.BlogInfo
	userAPI         handle.User
	pageAPI         handle.Page // 页面
	menuAPI         handle.Menu
	roleAPI         handle.Role         // 角色
	categoryAPI     handle.Category     // 分类
	tagAPI          handle.Tag          // 标签
	articleAPI      handle.Article      // 文章
	commentAPI      handle.Comment      // 评论
	messageAPI      handle.Message      // 留言
	linkAPI         handle.Link         // 友链
	resourceAPI     handle.Resource     // 资源
	operationLogAPI handle.OperationLog // 操作日志
	uploadAPI       handle.Upload       // 文件上传

	// 前台
	frontAPI handle.Front // 博客前台接口
)

func RegisterHandlers(r *gin.Engine) {
	// swagger配置：设置Swagger API文档的基础路径
	//SwaggerInfo是由 `swaggo/swag` 生成的文档信息配置结构体
	docs.SwaggerInfo.BasePath = "/api" // 设置 Swagger 文档的基础路径为 "/api"
	// 注册 Swagger UI 路由，`ginSwagger.WrapHandler` 会将 Swagger 文档与 UI 渲染结合
	// 这样用户可以通过访问 `/swagger` 路由来查看 API 文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 调用 registerBaseHandler 注册其他基础的路由处理
	// 例如，可能是一些常见的基础 API 路由
	registerBaseHandler(r)
	registerAdminHandler(r)
	registerBlogHandler(r)
}

// 通用接口: 全部不需要 登录 + 鉴权
func registerBaseHandler(r *gin.Engine) {
	base := r.Group("/api")

	// TODO: 登录, 注册 记录日志
	base.POST("/login", userAuthAPI.Login)            // 登录
	base.POST("/register", userAuthAPI.Register)      // 注册
	base.GET("/email/verify", userAuthAPI.VerifyCode) // 邮箱验证
	base.GET("/logout", userAuthAPI.Logout)           // 登出
	// TODO: 博客信息
	base.POST("/report", blogInfoAPI.Report)
	base.GET("/config", blogInfoAPI.GetConfigMap)
	base.PATCH("/config", blogInfoAPI.UpdateConfig)
}

// 后台管理系统的接口: 全部需要 登录 + 鉴权
func registerAdminHandler(r *gin.Engine) {
	auth := r.Group("/api")

	// 注意使用中间件的顺序
	auth.Use(middleware.JWTAuth())
	auth.Use(middleware.PermissionCheck())
	auth.Use(middleware.OperationLog())
	auth.Use(middleware.ListenOnline())

	auth.GET("/home", blogInfoAPI.GetHomeInfo)
	auth.POST("/upload", uploadAPI.UploadFile) // 文件上传

	// 用户模块
	user := auth.Group("/user")
	{
		user.GET("/info", userAPI.GetInfo)                           // 获取当前用户信息
		user.GET("/current", userAPI.UpdateCurrent)                  // 修改当前用户信息
		user.GET("/list", userAPI.GetList)                           // 用户列表
		user.PUT("", userAPI.Update)                                 // 更新用户信息
		user.PUT("/disable", userAPI.UpdateDisable)                  // 修改用户禁用状态
		user.PUT("/current/password", userAPI.UpdateCurrentPassword) // 修改当前用户密码
		user.GET("/online", userAPI.GetOnlineList)                   // 获取在线用户
		user.POST("/offline/:id", userAPI.ForceOffline)              // 强制用户下线
	}
	//博客设置
	setting := auth.Group("/setting")
	{
		setting.GET("/about", blogInfoAPI.GetAbout)    // 获取关于我
		setting.PUT("/about", blogInfoAPI.UpdateAbout) // 更新关于我
	}
	// 菜单模块
	menu := auth.Group("/menu")
	{
		menu.GET("/list", menuAPI.GetTreeList)      // 菜单列表
		menu.POST("", menuAPI.SaveOrUpdate)         // 新增/编辑菜单
		menu.DELETE("/:id", menuAPI.Delete)         // 删除菜单
		menu.GET("/user/list", menuAPI.GetUserMenu) // 获取当前用户的菜单
		menu.GET("/option", menuAPI.GetOption)      // 菜单选项列表(树形)
	}
	// 角色模块
	role := auth.Group("/role")
	{
		role.GET("/list", roleAPI.GetTreeList) // 角色列表(树形)
		role.POST("", roleAPI.SaveOrUpdate)    // 新增/编辑菜单
		role.DELETE("", roleAPI.Delete)        // 删除角色
		role.GET("/option", roleAPI.GetOption) // 角色选项列表(树形)
	}
	// 分类模块
	category := auth.Group("/category")
	{
		category.GET("/list", categoryAPI.GetList)     // 分类 列表
		category.POST("", categoryAPI.SaveOrUpdate)    // 新增/编辑分类
		category.DELETE("", categoryAPI.Delete)        // 删除分类
		category.GET("/option", categoryAPI.GetOption) // 分类选项列表
	}
	// 标签模块
	tag := auth.Group("/tag")
	{
		tag.GET("/list", tagAPI.GetList)     // 标签列表
		tag.POST("", tagAPI.SaveOrUpdate)    // 新增/编辑标签
		tag.DELETE("", tagAPI.Delete)        // 删除标签
		tag.GET("/option", tagAPI.GetOption) // 标签选项列表
	}
	// 文章模块
	articles := auth.Group("/article")
	{
		articles.GET("/list", articleAPI.GetList)                 // 文章列表
		articles.POST("", articleAPI.SaveOrUpdate)                // 新增/编辑文章
		articles.PUT("/top", articleAPI.UpdateTop)                // 更新文章置顶
		articles.GET("/:id", articleAPI.GetDetail)                // 文章详情
		articles.PUT("/soft-delete", articleAPI.UpdateSoftDelete) // 软删除文章
		articles.DELETE("", articleAPI.Delete)                    // 物理删除文章
		articles.POST("/export", articleAPI.Export)               // 导出文章
		articles.POST("/import", articleAPI.Import)               // 导入文章
	}
	// 评论模块
	comment := auth.Group("/comment")
	{
		comment.GET("/list", commentAPI.GetList)        // 评论列表
		comment.DELETE("", commentAPI.Delete)           // 删除评论
		comment.PUT("/review", commentAPI.UpdateReview) // 修改评论审核
	}
	// 留言模块
	message := auth.Group("/message")
	{
		message.GET("/list", messageAPI.GetList)        // 留言列表
		message.DELETE("", messageAPI.Delete)           // 删除留言
		message.PUT("/review", messageAPI.UpdateReview) // 审核留言
	}
	// 友情链接
	link := auth.Group("/link")
	{
		link.GET("/list", linkAPI.GetList)  // 友链列表
		link.POST("", linkAPI.SaveOrUpdate) // 新增/编辑友链
		link.DELETE("", linkAPI.Delete)     // 删除友链
	}
	// 资源模块
	resource := auth.Group("/resource")
	{
		resource.GET("/list", resourceAPI.GetTreeList)          // 资源列表(树形)
		resource.POST("", resourceAPI.SaveOrUpdate)             // 新增/编辑资源
		resource.DELETE("/:id", resourceAPI.Delete)             // 删除资源
		resource.PUT("/anonymous", resourceAPI.UpdateAnonymous) // 修改资源匿名访问
		resource.GET("/option", resourceAPI.GetOption)          // 资源选项列表(树形)
	}
	// 操作日志模块
	operationLog := auth.Group("/operation/log")
	{
		operationLog.GET("/list", operationLogAPI.GetList) // 操作日志列表
		operationLog.DELETE("", operationLogAPI.Delete)    // 删除操作日志
	}
	// 页面模块
	page := auth.Group("/page")
	{
		page.GET("/list", pageAPI.GetList)  // 页面列表
		page.POST("", pageAPI.SaveOrUpdate) // 新增/编辑页面
		page.DELETE("", pageAPI.Delete)     // 删除页面
	}
}

// 博客前台相关接口：大部分不需要登陆，部分需要登陆
func registerBlogHandler(r *gin.Engine) {
	base := r.Group("/api/front")

	base.GET("/about", blogInfoAPI.GetAbout) // 获取关于我
	base.GET("/home", frontAPI.GetHomeInfo)  //前台首页
	base.GET("/page", pageAPI.GetList)

	//需要登录
	base.Use(middleware.JWTAuth())
	{
		base.GET("/download/:id", uploadAPI.DownloadFile)
		base.HEAD("/download/:id", uploadAPI.DownloadFile) //文件下载
		base.POST("/upload", uploadAPI.UploadFile)         // 文件上传
		base.GET("/user/info", userAPI.GetInfo)            // 根据 Token 获取用户信息
		base.PUT("/user/info", userAPI.UpdateCurrent)      // 根据 Token 更新当前用户信息

		base.POST("/comment", frontAPI.SaveComment)                 // 前台新增评论
		base.GET("/comment/like/:comment_id", frontAPI.LikeComment) // 前台点赞评论
		base.POST("/message", frontAPI.SaveMessage)                 // 前台新增留言
		base.GET("/article/like/:article_id", frontAPI.LikeArticle) // 前台点赞文章
	}

	category := base.Group("/category")
	{
		category.GET("/list", frontAPI.GetCategoryList) // 前台分类列表
	}

	tag := base.Group("/tag")
	{
		tag.GET("/list", frontAPI.GetTagList) // 前台标签列表
	}

	article := base.Group("/article")
	{
		article.GET("/list", frontAPI.GetArticleList)    // 前台文章列表
		article.GET("/:id", frontAPI.GetArticleInfo)     // 前台文章详情
		article.GET("/archive", frontAPI.GetArchiveList) // 前台文章归档
		article.GET("/search", frontAPI.SearchArticle)   // 前台文章搜索
	}

	comment := base.Group("/comment")
	{
		comment.GET("/list", frontAPI.GetCommentList)                         // 前台评论列表
		comment.GET("/replies/:comment_id", frontAPI.GetReplyListByCommentId) // 根据评论 id 查询回复
	}

	message := base.Group("/message")
	{
		message.GET("/list", frontAPI.GetMessageList) // 前台留言列表
	}

	link := base.Group("/link")
	{
		link.GET("/list", frontAPI.GetLinkList) // 前台友链列表
	}

}
