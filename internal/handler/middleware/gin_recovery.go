package middleware

import (
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"

	"digital-wallet/platform/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RecoveryWithZap(logger logger.Logger, stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				if brokenPipe := isBrokenPipeError(err); brokenPipe {
					handleBrokenPipe(logger, c, err)
					return
				}

				httpRequest, er := httputil.DumpRequest(c.Request, false)
				if er != nil {
					logRequestError(logger, c, err, "unable to dump http request", "")
				}

				logMessage := "[Recovery from panic]"
				if stack {
					logMessage += "\n" + string(debug.Stack())
				}

				logRequestError(logger, c, err, logMessage, string(httpRequest))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"ok": false,
					"error": gin.H{
						"code":    http.StatusInternalServerError,
						"message": "Unexpected Internal Server Error. Please contact the administrator.",
					},
				})
			}
		}()
		c.Next()
	}
}

func isBrokenPipeError(err interface{}) bool {
	if ne, ok := err.(*net.OpError); ok {
		if se, ok := ne.Err.(*os.SyscallError); ok {
			return strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
				strings.Contains(strings.ToLower(se.Error()), "connection reset by peer")
		}
	}
	return false
}

func handleBrokenPipe(logger logger.Logger, c *gin.Context, err interface{}) {
	httpRequest, er := httputil.DumpRequest(c.Request, false)
	if er != nil {
		logRequestError(logger, c, err, "unable to dump http request", "")
	}
	logger.Error(c, c.Request.URL.Path,
		zap.Any("error", err),
		zap.String("request", string(httpRequest)),
	)
	_ = c.Error(err.(error))
	c.Abort()
}

func logRequestError(logger logger.Logger, c *gin.Context, err interface{}, msg, reqDump string) {
	logger.Error(c, msg,
		zap.Any("error", err),
		zap.String("request", string(reqDump)),
	)
}
