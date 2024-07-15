package controllers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/momokii/gin-crud-boilerplate/db"
	"github.com/momokii/gin-crud-boilerplate/models"
	"github.com/momokii/gin-crud-boilerplate/utils"
	"golang.org/x/crypto/bcrypt"
)

func GetSelf(c *gin.Context) {
	var user models.UserModelRes

	userData, exist := c.Get("user")
	if !exist {
		utils.ThrowErr(c, http.StatusUnauthorized, "User not found")
		return
	}

	user.Id = userData.(models.UserModel).Id
	user.Username = userData.(models.UserModel).Username
	user.Name = userData.(models.UserModel).Name
	user.Role = userData.(models.UserModel).Role
	user.IsActive = userData.(models.UserModel).IsActive

	c.JSON(http.StatusOK, gin.H{
		"errors":  false,
		"message": "Success Get All Users",
		"data":    user,
	})
}

func GetAllUsers(c *gin.Context) {
	var users []models.UserModelRes
	var is_active bool
	var total_user int

	// pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	per_page, _ := strconv.Atoi(c.DefaultQuery("per_page", "1"))
	offset := (page - 1) * per_page
	search := c.DefaultQuery("search", "")
	user_type := c.DefaultQuery("user_type", "")
	is_active_q := c.DefaultQuery("is_active", "")

	if (is_active_q == "1") || (is_active_q == "true") {
		is_active = true
	} else {
		is_active = false
	}

	query := "select id, username, name, role, is_active from users where 1=1"

	if search != "" {
		query += " and (username like '%" + search + "%' or name like '%" + search + "%')"
	}
	if user_type != "" {
		query += " and role = '" + user_type + "'"
	}

	query += " and is_active = " + strconv.FormatBool(is_active)

	if err := db.DB.QueryRow("select count(id) from (" + query + ") as total_user").Scan(&total_user); err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}

	userRows, err := db.DB.Query(query+" limit $1 offset $2", per_page, offset)
	if err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer userRows.Close()

	for userRows.Next() {
		var user models.UserModelRes
		if err := userRows.Scan(&user.Id, &user.Username, &user.Name, &user.Role, &user.IsActive); err != nil {
			utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
			return
		}
		users = append(users, user)
	}

	if users == nil {
		users = []models.UserModelRes{}
	}

	c.JSON(http.StatusOK, gin.H{
		"errors":  false,
		"message": "Success Get All Users",
		"data": gin.H{
			"page":       page,
			"per_page":   per_page,
			"total_data": total_user,
			"users":      users,
		},
	})
}

func GetOneUser(c *gin.Context) {
	var user models.UserModelRes

	user_id := c.Params.ByName("id")

	if err := db.DB.QueryRow("select id, username, name, role, is_active from users where id = $1", user_id).Scan(&user.Id, &user.Username, &user.Name, &user.Role, &user.IsActive); err != nil {
		if err == sql.ErrNoRows {
			utils.ThrowErr(c, http.StatusNotFound, "User not found")
		} else {
			utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errors":  false,
		"message": "Success Get Data User",
		"data":    user,
	})
}

func CreateUser(c *gin.Context) {
	type CreateUserInput struct {
		Username string `json:"username" binding:"required,alphanum,min-5"`
		Password string `json:"password" binding:"required"`
		Name     string `json:"name" binding:"required"`
		Role     string `json:"role" binding:"required"`
	}

	var inputUser CreateUserInput
	var user models.UserModel
	var tx *sql.Tx

	err := c.ShouldBindJSON(&inputUser)
	if err != nil {
		utils.ThrowErr(c, http.StatusBadRequest, "Invalid Request")
		return
	}

	validatePassword := utils.PasswordValidator(inputUser.Password)
	if !validatePassword {
		utils.ThrowErr(c, http.StatusBadRequest, "Password must be at least 6 characters and contain at least 1 uppercase letter and 1 number")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(inputUser.Password), 16)
	if err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}

	defer func() {
		if tx != nil {
			if err != nil {
				tx.Rollback()
			}
		}
	}()

	tx, err = db.DB.BeginTx(c, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}

	err = tx.QueryRow("select id from users where username = $1", inputUser.Username).Scan(&user.Id)
	if (err != nil) && (err != sql.ErrNoRows) {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}

	if user.Id != 0 {
		utils.ThrowErr(c, http.StatusBadRequest, "Username already exist")
		return
	}

	_, err = tx.Exec("insert into users (username, password, name, role, is_active, created_at, updated_at) values ($1, $2, $3, $4, true, now(), now())", inputUser.Username, string(hashedPassword), inputUser.Name, inputUser.Role, true)
	if err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err = tx.Commit(); err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errors":  false,
		"message": "Success Create User",
	})
}

func EditUser(c *gin.Context) {
	type EditUserInput struct {
		UserId int    `json:"user_id" binding:"required"`
		Name   string `json:"name" binding:"required"`
		Role   *int   `json:"role" binding:"required"`
	}

	var user models.UserModel
	var inputUser EditUserInput
	var tx *sql.Tx

	reqUser, exist := c.Get("user")
	if !exist {
		utils.ThrowErr(c, http.StatusUnauthorized, "User not found")
		return
	}

	err := c.ShouldBindJSON(&inputUser)
	if err != nil {
		utils.ThrowErr(c, http.StatusBadRequest, "Invalid Request")
		return
	}

	if reqUser.(models.UserModel).Role != 1 {
		inputUser.UserId = reqUser.(models.UserModel).Id
	}

	if (reqUser.(models.UserModel).Role != 1) && (inputUser.UserId != reqUser.(models.UserModel).Id) {
		utils.ThrowErr(c, http.StatusForbidden, "You can't edit another user (only admin can do that)")
		return
	}

	if (reqUser.(models.UserModel).Role != 1) && (inputUser.Role != nil) {
		utils.ThrowErr(c, http.StatusForbidden, "You can't edit role (only admin can do that)")
		return
	}

	if (reqUser.(models.UserModel).Role == 1 && inputUser.Role != nil) && (inputUser.UserId == reqUser.(models.UserModel).Id) {
		utils.ThrowErr(c, http.StatusForbidden, "You can't edit your own role")
		return
	}

	defer func() {
		if tx != nil {
			if err != nil {
				tx.Rollback()
			}
		}
	}()

	tx, err = db.DB.BeginTx(c, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}

	err = tx.QueryRow("select id from users where id = $1", inputUser.UserId).Scan(&user.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.ThrowErr(c, http.StatusNotFound, "User not found")
		} else {
			utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	data := []interface{}{inputUser.Name}
	paramIndex := 1
	query := fmt.Sprintf("update users set name = $%d", paramIndex)

	if inputUser.Role != nil {
		paramIndex++
		data = append(data, *inputUser.Role)
		query += fmt.Sprintf(", role = $%d", paramIndex)
	}

	paramIndex++
	query += fmt.Sprintf(" where id = $%d", paramIndex)
	data = append(data, inputUser.UserId)

	if _, err := tx.Exec(query, data...); err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err = tx.Commit(); err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errors":  false,
		"message": "Success Edit User",
	})
}

func EditUserPassword(c *gin.Context) {
	type ChangePasswordInput struct {
		PasswordNow string `json:"password_now" binding:"required"`
		PasswordNew string `json:"password_new" binding:"required"`
	}

	var inputUser ChangePasswordInput
	var tx *sql.Tx

	reqUser, exist := c.Get("user")
	if !exist {
		utils.ThrowErr(c, http.StatusUnauthorized, "User not found")
		return
	}

	err := c.ShouldBindJSON(&inputUser)
	if err != nil {
		utils.ThrowErr(c, http.StatusBadRequest, "Invalid Request")
		return
	}

	validatePassword := utils.PasswordValidator(inputUser.PasswordNew)
	if !validatePassword {
		utils.ThrowErr(c, http.StatusBadRequest, "Password must be at least 6 characters and contain at least 1 uppercase letter and 1 number")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(reqUser.(models.UserModel).Password), []byte(inputUser.PasswordNow)); err != nil {
		utils.ThrowErr(c, http.StatusUnauthorized, "Password Now is wrong")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(inputUser.PasswordNew), bcrypt.DefaultCost)
	if err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}

	defer func() {
		if tx != nil {
			if err != nil {
				tx.Rollback()
			}
		}
	}()

	tx, err = db.DB.BeginTx(c, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = tx.Exec("update users set password = $1 where id = $2", string(hashedPassword), reqUser.(models.UserModel).Id)
	if err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err = tx.Commit(); err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errors":  false,
		"message": "Success Edit Password User",
	})
}

func EditUserStatus(c *gin.Context) {
	id_user := c.Params.ByName("id")

	var tx *sql.Tx
	var user models.UserModel
	var err error

	reqUser, exist := c.Get("user")
	if !exist {
		utils.ThrowErr(c, http.StatusUnauthorized, "User not found")
		return
	}

	defer func() {
		if tx != nil {
			if err != nil {
				tx.Rollback()
			}
		}
	}()

	tx, err = db.DB.BeginTx(c, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}

	err = tx.QueryRow("select id, is_active from users where id = $1", id_user).Scan(&user.Id, &user.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.ThrowErr(c, http.StatusNotFound, "User not found")
		} else {
			utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	if reqUser.(models.UserModel).Id == user.Id {
		utils.ThrowErr(c, http.StatusForbidden, "You can't change your own status")
		return
	}

	_, err = tx.Exec("update users set is_active = not is_active where id = $1", id_user)
	if err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err = tx.Commit(); err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errors":  false,
		"message": "Success Edit User Status",
	})
}

func DeleteUser(c *gin.Context) {
	deleted_id := c.Params.ByName("id")

	var user models.UserModel
	var tx *sql.Tx
	var err error

	reqUser, exist := c.Get("user")
	if !exist {
		utils.ThrowErr(c, http.StatusUnauthorized, "User not found")
		return
	}

	defer func() {
		if tx != nil {
			if err != nil {
				tx.Rollback()
			}
		}
	}()

	tx, err = db.DB.BeginTx(c, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}

	err = tx.QueryRow("select id from users where id = $1", deleted_id).Scan(&user.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.ThrowErr(c, http.StatusNotFound, "User not found")
		} else {
			utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	if user.Id == reqUser.(models.UserModel).Id {
		utils.ThrowErr(c, http.StatusForbidden, "You can't delete yourself")
		return
	}

	_, err = tx.Exec("delete from users where id = $1", deleted_id)
	if err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err = tx.Commit(); err != nil {
		utils.ThrowErr(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errors":  false,
		"message": "Success Delete User",
	})
}
