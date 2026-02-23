package service

import (
	"errors"
	"polypus-project/common"
	"polypus-project/logger"
	"polypus-project/model"

	"gorm.io/gorm"
)

type User struct{}

type UserInfo struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,alphanum"`
}
type UserUpdate struct {
	ID          int64
	NewPassword string `json:"newPassword" binding:"required,alphanum"`
}
type UserShow struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Level    int
}

// 通用登录
func (s *User) Login(info UserInfo) (resp UserShow, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	//查询用户名
	var oneDoctor model.Doctor
	if err := tx.Model(&model.Doctor{}).Where("username = ?", info.Username).First(&oneDoctor).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return UserShow{}, common.ErrNew(errors.New("用户不存在"), common.ParamErr)
		} else {
			tx.Rollback()
			return UserShow{}, common.ErrNew(errors.New("用户名查询错误"), common.SysErr)
		}
	}
	//验证密码
	password := oneDoctor.Password
	if password != info.Password {
		tx.Rollback()
		return UserShow{}, common.ErrNew(errors.New("密码错误"), common.ParamErr)
	}
	//验证通过，返回用户信息
	resp = UserShow{
		ID:       oneDoctor.ID,
		Username: oneDoctor.Username,
		Level:    oneDoctor.Level,
	}
	if err := tx.Commit().Error; err != nil {
		return UserShow{}, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}
	return resp, nil
}

// 通用获取用户状态
func (u *User) Status(id int64) (resp UserShow, err error) {
	//查询用户信息
	var thisUser model.Doctor
	if err := model.DB.Model(&model.Doctor{}).Where("ID = ?", id).First(&thisUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Infof("已有用户信息因未知原因丢失")
			return UserShow{}, common.ErrNew(errors.New("账号不存在"), common.ParamErr)
		}
		return UserShow{}, common.ErrNew(errors.New("账号查找失败"), common.SysErr)
	}
	//返回对应用户信息
	resp = UserShow{
		ID:       thisUser.ID,
		Username: thisUser.Username,
		Level:    thisUser.Level,
	}
	return resp, nil
}

// 通用更新密码
func (s *User) Update(info UserUpdate) (err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	var thisDoctor model.Doctor
	if err := tx.Model(&model.Doctor{}).Where("id = ?", info.ID).First(&thisDoctor).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return common.ErrNew(errors.New("医生不存在"), common.ParamErr)
		}
		tx.Rollback()
		return common.ErrNew(errors.New("医生查询错误"), common.SysErr)
	}
	if thisDoctor.Level >= 2 {
		thisDoctor.Password = info.NewPassword
		if err := tx.Save(&thisDoctor).Error; err != nil {
			tx.Rollback()
			return common.ErrNew(errors.New("修改密码错误"), common.SysErr)
		}
	}
	if thisDoctor.Level == 1 {
		if thisDoctor.Password == "123456" {
			thisDoctor.Password = info.NewPassword
			if err := tx.Save(&thisDoctor).Error; err != nil {
				tx.Rollback()
				return common.ErrNew(errors.New("修改密码错误"), common.SysErr)
			}
		} else {
			tx.Rollback()
			return common.ErrNew(errors.New("请让超级管理员重置密码后修改"), common.ParamErr)
		}
	}
	if err := tx.Commit().Error; err != nil {
		return common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}
	return nil
}
