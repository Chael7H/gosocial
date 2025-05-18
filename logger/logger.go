package logger

import (
	"gosocial/settings"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Init  初始化Logger
func Init(cfg *settings.LogConfig) (err error) {
	writeSyncer := getLogWriter(
		cfg.FileName,
		cfg.MaxSize,
		cfg.MaxAge,
		cfg.MaxBackups,
	)
	encoder := getEncoder()
	var l = new(zapcore.Level)
	err = l.UnmarshalText([]byte(viper.GetString("log.level")))
	if err != nil {
		return
	}
	core := zapcore.NewCore(encoder, writeSyncer, l)
	lg := zap.New(core, zap.AddCaller())
	//替换zap库中全局的Logger
	zap.ReplaceGlobals(lg)
	return
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

// getLogWriter 日志记录器
func getLogWriter(filename string, maxSize, maxBackups, maxAge int) zapcore.WriteSyncer {
	//lumberJackLogger := &lumberjack.Logger{
	//	Filename:   filename,
	//	MaxSize:    maxSize,
	//	MaxBackups: maxBackups,
	//	MaxAge:     maxAge,
	//}
	//return zapcore.AddSync(lumberJackLogger)
	file, _ := os.Create("./test.log")
	// 利用io.MultiWriter支持文件和终端两个输出目标
	ws := io.MultiWriter(file, os.Stdout)
	return zapcore.AddSync(ws)
}

// GinLogger 接收Gin框架默认的日志
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()
		cost := time.Since(start)
		zap.L().Info(path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("ip", c.ClientIP()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.String("cost", cost.String()),
		)
	}
}

// GinRecovery recover掉项目可能出现的panic，并使用zap记录相关日志
func GinRecovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				//Check for a broken connection,as it is not really a
				//condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(net.Error); ok {
						if se.Temporary() {
							brokenPipe = true
						}
					}
				}
				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					zap.L().Error("http request error",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)))
					//if the connection is dead,we can't write a status to it.
					c.Error(err.(error)) //nolint: errcheck
					c.Abort()
					return
				}
				if stack {
					zap.L().Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					zap.L().Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
