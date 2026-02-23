package controller

import (
	"errors"
	"net/http"
	"polypus-project/common"
	"polypus-project/logger"
	"polypus-project/service"

	"github.com/gin-gonic/gin"
)

type User struct {
}

// 通用登录
func (s *User) Login(c *gin.Context) {
	var info service.UserInfo
	if err := c.ShouldBind(&info); err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(common.ErrNew(errors.New("输入参数无法解析"), common.ParamErr))
		return
	}
	resp, err := srv.User.Login(info)
	if err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(err)
		return
	}
	SessionSet(c, "user", UserSession{
		ID:       resp.ID,
		Username: resp.Username,
		Level:    resp.Level,
	})
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// 通用登出
func (u *User) Logout(c *gin.Context) {
	SessionClear(c)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// 通用获取用户状态
func (u *User) Status(c *gin.Context) {
	ID := SessionGet(c, "user").(UserSession).ID
	resp, err := srv.User.Status(ID)
	if err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// 通用修改密码
func (s *User) Update(c *gin.Context) {
	var info service.UserUpdate
	if err := c.ShouldBind(&info); err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(common.ErrNew(errors.New("输入参数无法解析"), common.ParamErr))
		return
	}
	info.ID = SessionGet(c, "user").(UserSession).ID
	err := srv.User.Update(info)
	if err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, gin.H{"message": "修改成功"}))
}
