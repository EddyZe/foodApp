package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Logger(log *logrus.Entry) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		host := c.Request.Host
		log.Infoln("Request: ", host, method, path, query)
		c.Next()

		status := c.Writer.Status()
		log.Infoln("Response: ", status, method, path, query)
	}
}
