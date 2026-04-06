package service

import (
	"errors"
	"polypus-project/common"
	"polypus-project/logger"
	"polypus-project/model"

	"github.com/alexedwards/argon2id"
	"github.com/sethvargo/go-password/password"
	"gorm.io/gorm"
)

type Admin struct{}

type DoctorInfo struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
}

type DoctorShow struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Phone    string `json:"phone"`
	Level    int    `json:"level"`
}

type DoctorShows struct {
	Total   int64        `json:"total"`
	Doctors []DoctorShow `json:"doctors"`
}
type DoctorReset struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	NewPassword string `json:"newPassword"`
}

// 超级管理员注册医生
func (s *Admin) DoctorRegister(info DoctorInfo) (resp DoctorShow, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	// //医生用户名验证
	// var existingUser []model.Doctor
	// if err := tx.Model(&model.Doctor{}).Where("username = ?", info.Username).Find(&existingUser).Error; err != nil {
	// 	tx.Rollback()
	// 	return DoctorShow{}, common.ErrNew(errors.New("已有用户查询失败"), common.SysErr)
	// }
	// if len(existingUser) > 0 {
	// 	tx.Rollback()
	// 	return DoctorShow{}, common.ErrNew(errors.New("用户名已存在"), common.ParamErr)
	// }
	//密码加密
	hash, err := argon2id.CreateHash(info.Password, argon2id.DefaultParams)
	if err != nil {
		tx.Rollback()
		return DoctorShow{}, common.ErrNew(errors.New("密码加密失败"), common.SysErr)
	}
	//提交医生信息
	doctor := &model.Doctor{
		Username: info.Username,
		Password: hash,
		Phone:    info.Phone,
		Level:    1, //默认医生权限
		Reset:    1, //默认可以重置密码
	}
	if err := tx.Model(&model.Doctor{}).Create(doctor).Error; err != nil {
		tx.Rollback()
		return DoctorShow{}, common.ErrNew(errors.New("医生创建错误"), common.SysErr)
	}
	resp = DoctorShow{
		ID:       doctor.ID,
		Username: doctor.Username,
		Phone:    doctor.Phone,
		Level:    doctor.Level,
	}
	if err := tx.Commit().Error; err != nil {
		return DoctorShow{}, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}

	return resp, nil
}

// 超级管理员提高医生权限
func (s *Admin) DoctorUpgrade(id int64) (resp DoctorShow, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	//查询医生是否存在
	var doctor model.Doctor
	if err := tx.Model(&model.Doctor{}).Where("id = ?", id).First(&doctor).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return DoctorShow{}, common.ErrNew(errors.New("账号不存在"), common.ParamErr)
		}
		tx.Rollback()
		return DoctorShow{}, common.ErrNew(errors.New("医生查询失败"), common.SysErr)
	}
	//提升至超级管理员权限
	newLevel := 2
	if err := tx.Model(&model.Doctor{}).Where("id = ?", id).Update("level", newLevel).Error; err != nil {
		tx.Rollback()
		return DoctorShow{}, common.ErrNew(errors.New("医生权限提升失败"), common.SysErr)
	}
	//返回新信息
	resp = DoctorShow{
		ID:       doctor.ID,
		Username: doctor.Username,
		Phone:    doctor.Phone,
		Level:    newLevel,
	}
	if err := tx.Commit().Error; err != nil {
		return DoctorShow{}, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}
	return resp, nil
}

// 超级管理员展示所有医生
func (s *Admin) DoctorShow(pagerForm common.PagerForm) (resp DoctorShows, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	//分页查询医生
	var doctors []model.Doctor
	var total int64
	query := model.DB.Model(&model.Doctor{})
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		logger.Infof("controller %v\n", err)
		return DoctorShows{}, common.ErrNew(errors.New("计算医生总数错误"), common.SysErr)
	}
	if err := query.Scopes(model.Paginate(pagerForm)).Find(&doctors).Error; err != nil {
		logger.Infof("controller %v\n", err)
		return DoctorShows{}, common.ErrNew(errors.New("查找医生错误"), common.SysErr)
	}
	resp.Total = total
	//返回医生信息
	resp.Doctors = make([]DoctorShow, len(doctors))
	for i, doctor := range doctors {
		resp.Doctors[i].ID = doctor.ID
		resp.Doctors[i].Username = doctor.Username
		resp.Doctors[i].Phone = doctor.Phone
		resp.Doctors[i].Level = doctor.Level
	}
	if err := tx.Commit().Error; err != nil {
		return DoctorShows{}, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}
	return resp, nil
}

// 超级管理员重置医生密码
func (s *Admin) DoctorReset(id int64, adminId int64) (resp DoctorReset, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	//查询医生是否存在
	var doctor model.Doctor
	if err := tx.Model(&model.Doctor{}).Where("id = ?", id).First(&doctor).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return DoctorReset{}, common.ErrNew(errors.New("账号不存在"), common.ParamErr)
		}
		tx.Rollback()
		return DoctorReset{}, common.ErrNew(errors.New("医生查询失败"), common.SysErr)
	}
	//验证权限
	if doctor.Level >= 2 && doctor.ID != adminId {
		tx.Rollback()
		return DoctorReset{}, common.ErrNew(errors.New("不可重置其他超级管理员的密码"), common.LevelErr)
	}
	//生成随机密码
	newPassword, err := password.Generate(16, 4, 0, false, false)
	//密码加密
	hash, err := argon2id.CreateHash(newPassword, argon2id.DefaultParams)
	if err != nil {
		tx.Rollback()
		return DoctorReset{}, common.ErrNew(errors.New("密码加密失败"), common.SysErr)
	}
	//更新密码和重置权限
	doctor.Password = hash
	doctor.Reset = 1
	if err := tx.Save(&doctor).Error; err != nil {
		tx.Rollback()
		return DoctorReset{}, common.ErrNew(errors.New("医生密码重置失败"), common.SysErr)
	}

	//返回新信息
	resp = DoctorReset{
		ID:          doctor.ID,
		Username:    doctor.Username,
		NewPassword: newPassword,
	}
	if err := tx.Commit().Error; err != nil {
		return DoctorReset{}, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}
	return resp, nil
}

// 超级管理员删除医生
func (s *Admin) DoctorDelete(id int64, adminId int64) (err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	//查询医生是否存在
	var doctor model.Doctor
	if err := tx.Model(&model.Doctor{}).Where("id = ?", id).First(&doctor).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return common.ErrNew(errors.New("账号不存在"), common.ParamErr)
		}
		tx.Rollback()
		return common.ErrNew(errors.New("医生查询失败"), common.SysErr)
	}
	//验证权限
	if doctor.Level >= 2 {
		tx.Rollback()
		return common.ErrNew(errors.New("不可删除超级管理员"), common.LevelErr)
	}
	newLevel := -1
	if err := tx.Model(&model.Doctor{}).Where("id = ?", id).Update("level", newLevel).Error; err != nil {
		tx.Rollback()
		return common.ErrNew(errors.New("删除医生失败"), common.SysErr)
	}
	if err := tx.Commit().Error; err != nil {
		return common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}
	return nil
}
