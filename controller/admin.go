package controller

import (
	"errors"
	"net/http"
	"polypus-project/common"
	"polypus-project/logger"
	"polypus-project/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 超级管理员
type Admin struct{}

// 超级管理员注册医生
func (s *Admin) DoctorRegister(c *gin.Context) {
	var info service.DoctorInfo
	if err := c.ShouldBind(&info); err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(common.ErrNew(errors.New("输入参数无法解析"), common.ParamErr))
		return
	}
	resp, err := srv.Admin.DoctorRegister(info)
	if err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// 超级管理员提高医生权限
func (s *Admin) DoctorUpgrade(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("doctorid"), 10, 64)
	if err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(common.ErrNew(errors.New("输入参数无法解析"), common.ParamErr))
		return
	}
	resp, err := srv.Admin.DoctorUpgrade(id)
	if err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// 超级管理员展示所有医生
func (s *Admin) DoctorShow(c *gin.Context) {
	var pagerForm common.PagerForm
	if err := c.ShouldBind(&pagerForm); err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(common.ErrNew(errors.New("输入参数无法解析"), common.ParamErr))
		return
	}
	resp, err := srv.Admin.DoctorShow(pagerForm)
	if err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// 超级管理员重置医生密码

func (s *Admin) DoctorReset(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("doctorid"), 10, 64)
	if err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(common.ErrNew(errors.New("输入参数无法解析"), common.ParamErr))
		return
	}
	adminId := SessionGet(c, "user").(UserSession).ID
	resp, err := srv.Admin.DoctorReset(id, adminId)
	if err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
