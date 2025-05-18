package routes

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	swaggerFiles "github.com/swaggo/files" // 新路径
	ginSwagger "github.com/swaggo/gin-swagger"
	"gosocial/controllers"
	"gosocial/dao/mysql"
	rd "gosocial/dao/redis"
	_ "gosocial/docs" // 导入生成的docs包
	"gosocial/settings"

	"gosocial/logger"
	"gosocial/middlewares"
	"net/http"
)

// Init  初始化路由连接
func Init() *gin.Engine {
	// 初始化Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", settings.Conf.RedisConfig.Host, settings.Conf.RedisConfig.Port),
		Password: settings.Conf.RedisConfig.Password,
		DB:       settings.Conf.RedisConfig.DB,
	})
	// 创建MessageDao
	messageDao := rd.NewMessageDao(redisClient)
	mysqlMessageDao := mysql.NewMessageDao()
	r := gin.New()
	r.Use(logger.GinLogger(), logger.GinRecovery(true))

	// 静态文件服务 - 统一处理所有静态资源
	r.Static("/static", "./static")
	// 配置模板
	r.GET("/", func(c *gin.Context) {
		c.File("./templates/index.html")
	})
	r.GET("/friends", func(c *gin.Context) {
		c.File("./templates/friends.html")
	})
	r.GET("/user/info", func(c *gin.Context) {
		c.File("./templates/user/info.html")
	})
	r.GET("/friends/:friendID", func(c *gin.Context) {
		c.File("./templates/friends/detail.html")
	})
	r.GET("/posts", func(c *gin.Context) {
		c.File("./templates/posts.html")
	})

	v1 := r.Group("/api/v1")
	v1.POST("/register", controllers.RegisterHandler) //注册业务路由
	v1.POST("/login", controllers.LoginHandler)       //登录业务路由

	// 初始化控制器
	messageCtrl := controllers.NewMessageController(messageDao, mysqlMessageDao)
	uploadCtrl := controllers.NewUploadController()

	v1.Use(middlewares.JWTAuthMiddleware())
	{
		// 个人中心相关路由
		v1.GET("/user/info", controllers.GetUserInfoHandler)               //获取个人信息
		v1.PUT("/user/update_password", controllers.UpdatePasswordHandler) //更新密码
		v1.PUT("/user/update_info", controllers.UpdateUserInfoHandler)     //更新用户信息
		v1.POST("/user/upload_avatar", controllers.UploadAvatarHandler)    //上传头像

		// 好友相关路由
		v1.POST("/friends/:friendID", controllers.AddFriendHandler)      //添加好友
		v1.GET("/friends", controllers.GetFriendListHandler)             //好友列表
		v1.GET("/friends/:friendID", controllers.GetFriendDetailHandler) //获取好友详情信息
		v1.GET("/friends/search", controllers.SearchFriendHandler)       //搜索好友
		v1.PUT("/friends/", controllers.UpdateFriendRemarkHandler)       //更新好友备注
		v1.DELETE("/friends", controllers.DeleteFriendHandler)           //删除好友

		// 消息相关路由
		v1.POST("/messages", messageCtrl.SendMessageHandler)            //发送文本消息
		v1.POST("/messages/image", messageCtrl.SendImageMessageHandler) //发送图片消息
		v1.POST("/messages/file", messageCtrl.SendFileMessageHandler)   //发送文件消息
		v1.GET("/messages", messageCtrl.GetMessagesHandler)             //获取消息记录
		v1.GET("/messages/unread", messageCtrl.GetUnreadCountsHandler)  //获取未读消息数

		// 上传路由
		v1.POST("/upload", uploadCtrl.UploadFileHandler) // 文件上传

		// 动态相关路由
		v1.POST("/posts", controllers.CreatePostHandler)                //用户创建动态
		v1.GET("/posts", controllers.GetFriendPostsHandler)             //获取所有好友动态列表(个人空间)
		v1.GET("/posts/:uid", controllers.GetUserPostsHandler)          //获取指定用户动态列表
		v1.PUT("/posts/:id/view", controllers.IncrementPostViewHandler) //增加动态浏览量
		v1.DELETE("/posts/:id", controllers.DeletePostHandler)          //用户删除动态
	}

	// 添加Swagger路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 404处理
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "404 Page Not Found",
		})
	})

	return r
}
