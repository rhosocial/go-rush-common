package error

import (
	"net/http"

	"github.com/gin-gonic/gin"
	response2 "github.com/rhosocial/go-rush-common/component/response"
)

// ErrorHandler 定义一个中间件，用于捕获错误并统一返回
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用 defer 来捕获 panic
		defer func() {
			if err := recover(); err != nil {
				// 构造一个错误响应
				response := response2.Generic[interface{}, interface{}]{
					Base: response2.Base{
						Code:    http.StatusInternalServerError,
						Message: http.StatusText(http.StatusInternalServerError),
					},
				}

				// 返回错误响应
				c.JSON(http.StatusInternalServerError, response)
			}
		}()

		// 继续处理请求
		c.Next()
	}
}
