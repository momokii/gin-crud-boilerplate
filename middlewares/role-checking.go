package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/momokii/gin-crud-boilerplate/utils"
)

func IsAdmin(c *gin.Context) {
	role, exist := c.Get("role")
	if !exist {
		utils.ThrowErr(c, http.StatusUnauthorized, "Role not found")
		c.Abort()
		return
	}

	if role != 1 {
		utils.ThrowErr(c, http.StatusUnauthorized, "You are not authorized (not admin)")
		c.Abort()
		return
	}

	c.Next()
}
