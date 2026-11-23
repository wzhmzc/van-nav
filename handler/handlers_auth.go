package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mereith/nav/database"
	"github.com/mereith/nav/logger"
	"github.com/mereith/nav/service"
	"github.com/mereith/nav/types"
	"github.com/mereith/nav/utils"
)

func DeleteApiTokenHandler(c *gin.Context) {
	// 删除 Token
	id := c.Param("id")
	err := service.DeleteApiToken(id)
	if err != nil {
		utils.CheckErr(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"success": true,
		"message": "删除 API Token 成功",
	})
}

func AddApiTokenHandler(c *gin.Context) {
	var token types.AddTokenDto
	err := c.ShouldBindJSON(&token)
	if err != nil {
		utils.CheckErr(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}
	newId := utils.GenerateId()
	var signedJwt string
	signedJwt, err = utils.SignJWTForAPI(token.Name, newId)
	if err != nil {
		utils.CheckErr(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}
	service.AddApiTokenInDB(types.Token{
		Name:     token.Name,
		Value:    signedJwt,
		Id:       newId,
		Disabled: 0,
	})
	// 签名 jwt
	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"id":    newId,
			"Value": signedJwt,
			"Name":  token.Name,
		},
		"message": "添加 Token 成功",
	})
}

func LoginHandler(c *gin.Context) {
	var data types.LoginDto
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.CheckErr(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"errorMessage": err.Error(),
		})
		return
	}
	user, err := service.GetUser(data.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "errorMessage": err.Error()})
		return
	}
	if user.Name == "" {
		c.JSON(200, gin.H{
			"success":      false,
			"errorMessage": "用户名不存在",
		})
		return
	}
	// 使用 bcrypt 验证密码（兼容旧版明文：自动升级为哈希）
	if !utils.CheckPassword(data.Password, user.Password) {
		// 兜底：旧版明文密码兼容（登录成功后自动升级为 bcrypt 哈希）
		if user.Password != data.Password {
			c.JSON(200, gin.H{
				"success":      false,
				"errorMessage": "密码错误",
			})
			return
		}
		// 明文匹配 → 自动升级为 bcrypt 哈希
		if hashed, err := utils.HashPassword(data.Password); err == nil {
			service.UpgradeUserPassword(user.Id, hashed)
			logger.LogInfo("用户 %s 密码已自动升级为 bcrypt 哈希", user.Name)
		}
	}
	// 生成 token（嵌入当前 token_version，密码修改后旧 token 失效）
	tokenVersion := database.GetUserTokenVersion(user.Id)
	token, err := utils.SignJWT(user, tokenVersion)
	utils.CheckErr(err)

	c.JSON(200, gin.H{
		"success": true,
		"message": "登录成功",
		"data": gin.H{
			"user":  user,
			"token": token,
		},
	})

}

// 退出登录

// 退出登录
func LogoutHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"success": true,
		"message": "登出成功",
	})
}
