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

// 病历单
type Form struct{}

func (s *Form) PatientRegister(c *gin.Context) {
	var name service.Patient
	if err := c.ShouldBind(&name); err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(common.ErrNew(errors.New("输入参数无法解析"), common.ParamErr))
		return
	}
	respid, err := srv.Form.PatientRegister(name)
	if err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, gin.H{"ClinicNumber": respid}))
}
func (s *Form) FormInput(c *gin.Context) {
	var info service.FormInfo
	if err := c.ShouldBind(&info); err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(common.ErrNew(errors.New("输入参数无法解析"), common.ParamErr))
		return
	}
	resp, err := srv.Form.FormInput(info)
	if err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
func (s *Form) FormSave(c *gin.Context) {
	var info service.FormSave
	if err := c.ShouldBind(&info); err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(common.ErrNew(errors.New("输入参数无法解析"), common.ParamErr))
		return
	}
	info.DoctorId = SessionGet(c, "user").(UserSession).ID
	resp, err := srv.Form.FormSave(info)
	if err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, gin.H{"formId": resp}))
}
func (s *Form) PatientHistory(c *gin.Context) {
	var info service.PatientHistoryGet
	if err := c.ShouldBind(&info); err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(common.ErrNew(errors.New("输入参数无法解析"), common.ParamErr))
		return
	}
	resp, err := srv.Form.PatientHistory(info)
	if err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
func (s *Form) FormUpdate(c *gin.Context) {
	var info service.FormUpdate
	if err := c.ShouldBind(&info); err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(common.ErrNew(errors.New("输入参数无法解析"), common.ParamErr))
		return
	}
	resp, err := srv.Form.FormUpdate(info)
	if err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
func (s *Form) GetAll(c *gin.Context) {
	var info service.FormGet
	if err := c.ShouldBind(&info); err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(common.ErrNew(errors.New("输入参数无法解析"), common.ParamErr))
		return
	}
	resp, err := srv.Form.GetAll(info)
	if err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

func (s *Form) FormDelete(c *gin.Context) {
	formId, err := strconv.ParseInt(c.Param("formid"), 10, 64)
	if err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(common.ErrNew(errors.New("输入参数无法解析"), common.ParamErr))
		return
	}
	err = srv.Form.FormDelete(formId)
	if err != nil {
		logger.Infof("controller %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, gin.H{"message": "删除成功"}))
}

// func (s *Form) Download(c *gin.Context) {
// 	var url string
// 	url = c.Param("url")
// 	// 安全检查：防止目录遍历攻击
// 	if strings.Contains(url, "..") || strings.Contains(url, "/") {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "非法文件名",
// 		})
// 		return
// 	}
// 	// 构建文件路径
// 	filePath := filepath.Join("csv", url)
// 	// 检查文件是否存在
// 	if _, err := os.Stat(filePath); os.IsNotExist(err) {
// 		c.JSON(http.StatusNotFound, gin.H{
// 			"error": "文件不存在",
// 		})
// 		return
// 	}
// 	// 设置响应头，触发下载
// 	c.Header("Content-Description", "File Transfer")
// 	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", url))
// 	c.Header("Content-Type", "application/octet-stream")
// 	c.Header("Content-Transfer-Encoding", "binary")
// 	c.Header("Expires", "0")
// 	c.Header("Cache-Control", "must-revalidate")
// 	c.Header("Pragma", "public")
// 	// 发送文件
// 	c.File(filePath)
// }
