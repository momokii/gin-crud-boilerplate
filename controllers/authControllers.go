package controllers

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/momokii/gin-crud-boilerplate/db"
	"github.com/momokii/gin-crud-boilerplate/models"
	"github.com/momokii/gin-crud-boilerplate/utils"
	"golang.org/x/crypto/bcrypt"
)

func Login(c *gin.Context) {
	type LoginData struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	var loginData LoginData
	var user models.UserModel

	err := c.ShouldBindJSON(&loginData)
	if err != nil {
		utils.ThrowErr(c, http.StatusBadRequest, "Invalid Request")
		return
	}

	if err = db.DB.QueryRow("select id, username, password, is_active from users where username = $1", loginData.Username).Scan(&user.Id, &user.Username, &user.Password, &user.IsActive); err != nil {
		if err == sql.ErrNoRows {
			utils.ThrowErr(c, http.StatusUnauthorized, "Username or Password is wrongd")
		} else {
			utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password)); err != nil {
		utils.ThrowErr(c, http.StatusUnauthorized, "Username or Password is wrong")
		return
	}

	if !user.IsActive {
		utils.ThrowErr(c, http.StatusUnauthorized, "Your account is not active")
		return
	}

	sign := jwt.New(jwt.SigningMethodHS256)
	claims := sign.Claims.(jwt.MapClaims)
	claims["userId"] = user.Id
	claims["exp"] = time.Now().Add(time.Hour * 24 * 30).Unix() // 30 days\

	token, err := sign.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errors":  false,
		"message": "Login Success",
		"data": gin.H{
			"token":   token,
			"type":    "Bearer",
			"expired": "30d",
		},
	})
}
