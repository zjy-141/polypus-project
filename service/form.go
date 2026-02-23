package service

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"polypus-project/common"
	"polypus-project/logger"
	"polypus-project/model"
	"strconv"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Form struct{}

type Patient struct {
	PatientNumber string `json:"PatientNumber" binding:"required"`
}
type FormInfo struct {
	Age           int     `form:"age" json:"age" binding:"required"`
	PolypsNumber  int     `form:"polypsNumber" json:"polyps" binding:"required,min=1"`
	LongDiameter  float64 `form:"longDiameter" json:"long_diameter" binding:"required,min=0"`
	ShortDiameter float64 `form:"shortDiameter" json:"short_diameter" binding:"required,min=0"`
	BaseType      int     `form:"baseType" json:"base" binding:"required,oneof=1 2"`
}
type ResponseData struct {
	Result FormResp `json:"result"`
}
type FormResp struct {
	Probability float64 `json:"probability"`
	RiskLever   string  `json:"risk_level"`
	RiskLevelCn string  `json:"risk_level_cn"`
	Advice      string  `json:"advice"`
}
type FormSave struct {
	DoctorId      int64
	PatientNumber string  `json:"PatientNumber" binding:"required"`
	Age           int     `json:"age" binding:"required"`
	PolypsNumber  int     `json:"polypsNumber" binding:"required,min=1"`
	LongDiameter  float64 `json:"longDiameter" binding:"required,min=0"`
	ShortDiameter float64 `json:"shortDiameter" binding:"required,min=0"`
	BaseType      int     `json:"baseType" binding:"required,oneof=1 2"`
	Probability   float64 `json:"probability" binding:"omitempty"`
	RiskLevel     string  `json:"riskLevel" binding:"omitempty"`
	Advice        string  `json:"advice" binding:"omitempty"`
}

type PatientHistoryGet struct {
	PatientNumber string `json:"PatientNumber" binding:"required"`
	common.PagerForm
}
type PatientForm struct {
	FormId        int64   `json:"formId"`
	DoctorId      int64   `json:"doctorId" binding:"required"`
	DoctorName    string  `json:"doctorName"`
	PatientNumber string  `json:"PatientNumber" binding:"required"`
	FormTime      string  `json:"formTime"`
	Age           int     `json:"age" binding:"required"`
	PolypsNumber  int     `json:"polypsNumber" binding:"required,min=1"`
	LongDiameter  float64 `json:"longDiameter" binding:"required,min=0"`
	ShortDiameter float64 `json:"shortDiameter" binding:"required,min=0"`
	BaseType      int     `json:"baseType" binding:"required,oneof=1 2"`
	Probability   float64 `json:"probability" binding:"required"`
	RiskLevel     string  `json:"risk_level" binding:"required"`
	Advice        string  `json:"advice" binding:"required"`
	IsWorse       string  `json:"is_worse"`
	Comment       string  `json:"comment"`
}
type PatientHistory struct {
	Total        int64         `json:"total"`
	PatientForms []PatientForm `json:"patientForms"`
}
type FormUpdate struct {
	FormId  int    `form:"formId" binding:"required"`
	IsWorse string `form:"isWorse" binding:"omitempty,oneof=-1 0 1"`
	Comment string `form:"comment" binding:"omitempty"`
}
type FormGet struct {
	DoctorId         int64   `form:"doctorId" binding:"omitempty"`
	PatientNumber    string  `form:"PatientNumber" binding:"omitempty"`
	AgeMin           int     `form:"age_min" binding:"omitempty"`
	AgeMax           int     `form:"age_max" binding:"omitempty"`
	PolypsNumberMin  int     `form:"polypsNumber_min" binding:"omitempty"`
	PolypsNumberMax  int     `form:"polypsNumber_max" binding:"omitempty"`
	LongDiameterMin  float64 `form:"longDiameter_min" binding:"omitempty,min=0"`
	LongDiameterMax  float64 `form:"longDiameter_max" binding:"omitempty,min=0"`
	ShortDiameterMax float64 `form:"shortDiameter_max" binding:"omitempty,min=0"`
	ShortDiameterMin float64 `form:"shortDiameter_min" binding:"omitempty,min=0"`
	BaseType         string  `form:"baseType" binding:"omitempty,oneof=1 2"`
	RiskLevel        string  `form:"riskLevel" binding:"omitempty"`
	Comment          string  `form:"comment"`
	common.PagerForm
}

type FormShow struct {
	Total int64         `json:"total"`
	Forms []PatientForm `json:"forms"`
	Url   string        `json:"url"`
}

func (s *Form) PatientRegister(info Patient) (resp string, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	if number, err := strconv.Atoi(info.PatientNumber); err != nil || number <= 0 {
		return "", common.ErrNew(errors.New("病历号不符合条件"), common.ParamErr)
	}
	var patient model.Patient
	patient.PatientNumber = info.PatientNumber
	if err := tx.Model(&model.Patient{}).Create(&patient).Error; err != nil {
		tx.Rollback()
		logger.Errorf("患者创建错误: %v", err)
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return info.PatientNumber, nil
		} else {
			return "", common.ErrNew(errors.New("患者创建错误"), common.SysErr)
		}
	}
	resp = patient.PatientNumber
	if err := tx.Commit().Error; err != nil {
		return "", common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}
	return resp, nil
}
func (s *Form) FormInput(info FormInfo) (resp FormResp, err error) {
	if info.LongDiameter < info.ShortDiameter {
		return FormResp{}, common.ErrNew(errors.New("长径不能小于短径"), common.SysErr)
	}
	//调用Python接口
	host := os.Getenv("PREDICT_HOST")
	if host == "" {
		host = "127.0.0.1" // 默认值
	}
	port := os.Getenv("PREDICT_PORT")
	if port == "" {
		port = "8087" // 默认值
	}
	// 构建 URL
	url := fmt.Sprintf("http://%s:%s/api/predict", host, port)
	jsonData, err := json.Marshal(info)
	if err != nil {
		return FormResp{}, err
	}
	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 30 * time.Second, //表示30秒超时
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return FormResp{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	pyresp, err := client.Do(req)
	if err != nil {
		return FormResp{}, err
	}
	// logger.Infof("---------------------------------------------%#v\n", pyresp)
	defer pyresp.Body.Close()
	body, err := io.ReadAll(pyresp.Body)
	if err != nil {
		return FormResp{}, err
	}
	// 解析JSON
	var data ResponseData
	if err := json.Unmarshal(body, &data); err != nil {
		return FormResp{}, err
	}
	// if resp1.StatusCode != 200 {
	// 	errMsg := service.UserPostError{}
	// 	if err := json.Unmarshal(bodyBytes, &errMsg); err != nil {
	// 		return FormResp{}, err
	// 	}
	resp = data.Result
	return resp, nil
}

func (s *Form) FormSave(info FormSave) (resp int64, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	var patient model.Patient
	if err := tx.Model(&model.Patient{}).Where("patient_number = ?", info.PatientNumber).First(&patient).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return -1, nil
		}
		tx.Rollback()
		return 0, common.ErrNew(errors.New("患者不存在"), common.LevelErr)
	}
	var form model.Form
	form.DoctorID = info.DoctorId
	form.PatientNumber = info.PatientNumber
	form.FormTime = time.Now().Format("2006-01-02 15:04:05")
	form.Age = info.Age
	form.LongDiameter = info.LongDiameter
	form.ShortDiameter = info.ShortDiameter
	form.BaseType = info.BaseType
	form.PolypNumber = info.PolypsNumber
	form.RiskLevel = info.RiskLevel
	form.Probability = info.Probability
	form.Advice = info.Advice
	if err := tx.Model(&model.Form{}).Create(&form).Error; err != nil {
		tx.Rollback()
		logger.Errorf("病历单创建错误: %v", err)
		return 0, common.ErrNew(errors.New("病历单创建错误"), common.SysErr)
	}
	resp = form.ID
	if err := tx.Commit().Error; err != nil {
		return 0, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}
	return resp, nil
}
func (s *Form) PatientHistory(get PatientHistoryGet) (resp PatientHistory, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	var forms []model.Form
	var total int64
	query := model.DB.Model(&model.Form{}).Where("patient_number like ?", get.PatientNumber+"%")

	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		logger.Infof("controller %v\n", err)
		return PatientHistory{}, common.ErrNew(errors.New("计算病历单总数错误"), common.SysErr)
	}
	if err := query.Scopes(model.Paginate(get.PagerForm)).Preload("Doctor").Find(&forms).Error; err != nil {
		logger.Infof("controller %v\n", err)
		return PatientHistory{}, common.ErrNew(errors.New("查找病历单错误"), common.SysErr)
	}
	// logger.Infof("-----------------------------------forms %#v\n", forms)
	resp.Total = total
	resp.PatientForms = make([]PatientForm, len(forms))
	for i, form := range forms {
		resp.PatientForms[i].FormId = form.ID
		resp.PatientForms[i].DoctorId = form.DoctorID
		resp.PatientForms[i].DoctorName = form.Doctor.Username
		resp.PatientForms[i].PatientNumber = form.PatientNumber
		resp.PatientForms[i].FormTime = form.FormTime
		resp.PatientForms[i].Age = form.Age
		resp.PatientForms[i].PolypsNumber = form.PolypNumber
		resp.PatientForms[i].LongDiameter = form.LongDiameter
		resp.PatientForms[i].ShortDiameter = form.ShortDiameter
		resp.PatientForms[i].BaseType = form.BaseType
		resp.PatientForms[i].Probability = form.Probability
		resp.PatientForms[i].RiskLevel = form.RiskLevel
		resp.PatientForms[i].Advice = form.Advice
		resp.PatientForms[i].IsWorse = form.IsWorse
		resp.PatientForms[i].Comment = form.Comment
	}
	if err := tx.Commit().Error; err != nil {
		return PatientHistory{}, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}
	return resp, nil

}
func (s *Form) FormUpdate(info FormUpdate) (resp int64, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	var form model.Form
	if err := tx.Model(&model.Form{}).Where("id = ?", info.FormId).First(&form).Error; err != nil {
		tx.Rollback()
		logger.Errorf("查找病历单错误: %v", err)
		return 0, common.ErrNew(errors.New("查找病历单错误"), common.SysErr)
	}
	if info.IsWorse != "" {
		form.IsWorse = info.IsWorse
	}
	if info.Comment != "" {
		form.Comment = info.Comment
	}
	if err := tx.Model(&model.Form{}).Where("id = ?", info.FormId).Updates(form).Error; err != nil {
		tx.Rollback()
		logger.Errorf("更新病历单错误: %v", err)
		return 0, common.ErrNew(errors.New("更新病历单错误"), common.SysErr)
	}
	resp = form.ID
	if err := tx.Commit().Error; err != nil {
		return 0, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}
	return resp, nil
}
func (s *Form) GetAll(get FormGet) (resp FormShow, err error) {
	var forms []model.Form
	var total int64
	query := model.DB.Model(&model.Form{})
	// logger.Infof("-----------------------------------query %#v\n", get)
	if get.DoctorId != 0 {
		query = query.Where("doctor_id = ?", get.DoctorId)
	}
	if get.PatientNumber != "" {
		query = query.Where("patient_number like ?", get.PatientNumber+"%")
	}
	if get.AgeMax != 0 {
		query = query.Where("age <= ?", get.AgeMax)
	}
	if get.AgeMin != 0 {
		query = query.Where("age >= ?", get.AgeMin)
	}
	if get.PolypsNumberMax != 0 {
		query = query.Where("polyp_number <= ?", get.PolypsNumberMax)
	}
	if get.PolypsNumberMin != 0 {
		query = query.Where("polyp_number >= ?", get.PolypsNumberMin)
	}
	if get.LongDiameterMax != 0 {
		query = query.Where("long_diameter <= ?", get.LongDiameterMax)
	}
	if get.LongDiameterMin != 0 {
		query = query.Where("long_diameter >= ?", get.LongDiameterMin)
	}
	if get.ShortDiameterMax != 0 {
		query = query.Where("short_diameter <= ?", get.ShortDiameterMax)
	}
	if get.ShortDiameterMin != 0 {
		query = query.Where("short_diameter >= ?", get.ShortDiameterMin)
	}
	if get.BaseType != "" {
		query = query.Where("base_type = ?", get.BaseType)
	}
	if get.RiskLevel != "" {
		query = query.Where("risk_level = ?", get.RiskLevel)
	}
	if get.Comment != "" {
		query = query.Where("comment like ?", "%"+get.Comment+"%")
	}
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		logger.Infof("controller %v\n", err)
		return FormShow{}, common.ErrNew(errors.New("计算委托总数错误"), common.SysErr)
	}
	if err := query.Scopes(model.Paginate(get.PagerForm)).Preload("Doctor").Find(&forms).Error; err != nil {
		return FormShow{}, common.ErrNew(errors.New("查找委托错误"), common.SysErr)
	}
	resp.Total = total
	resp.Forms = make([]PatientForm, len(forms))
	for i, form := range forms {
		resp.Forms[i].FormId = form.ID
		resp.Forms[i].DoctorId = form.DoctorID
		resp.Forms[i].DoctorName = form.Doctor.Username
		resp.Forms[i].PatientNumber = form.PatientNumber
		resp.Forms[i].FormTime = form.FormTime
		resp.Forms[i].Age = form.Age
		resp.Forms[i].PolypsNumber = form.PolypNumber
		resp.Forms[i].LongDiameter = form.LongDiameter
		resp.Forms[i].ShortDiameter = form.ShortDiameter
		resp.Forms[i].BaseType = form.BaseType
		resp.Forms[i].Probability = form.Probability
		resp.Forms[i].RiskLevel = form.RiskLevel
		resp.Forms[i].Advice = form.Advice
		resp.Forms[i].IsWorse = form.IsWorse
		resp.Forms[i].Comment = form.Comment
	}
	// 添加CSV导出功能
	cwd, err := os.Getwd()
	if err != nil {
		return FormShow{}, err
	}
	csvDir := filepath.Join(cwd, "csv")

	// 确保目录存在
	if err := os.MkdirAll(csvDir, 0755); err != nil {
		return FormShow{}, err
	}
	filename := uuid.New().String() + ".csv"
	// 完整的文件路径
	filePath := filepath.Join(csvDir, filename)

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return FormShow{}, err
	}
	defer file.Close()
	// 关键：写入UTF-8 BOM（解决Excel中文乱码）
	bom := []byte{0xEF, 0xBB, 0xBF}
	if _, err := file.Write(bom); err != nil {
		return FormShow{}, err
	}

	// 创建 CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	header := []string{"FormID", "DoctorID", "PatientNumber", "FormTime", "Age", "LongDiameter", "ShortDiameter", "BaseType", "PolypsNumber", "RiskLevel", "Probability", "Advice", "Comment"}
	if err := writer.Write(header); err != nil {
		return FormShow{}, err
	}
	// 写入数据行
	for _, form := range forms {
		record := []string{
			strconv.FormatInt(form.ID, 10),
			strconv.FormatInt(form.DoctorID, 10),
			form.PatientNumber,
			form.FormTime,
			strconv.Itoa(form.Age),
			strconv.FormatFloat(form.LongDiameter, 'f', 2, 64),
			strconv.FormatFloat(form.ShortDiameter, 'f', 2, 64),
			strconv.Itoa(form.BaseType),
			strconv.Itoa(form.PolypNumber),
			form.RiskLevel,
			strconv.FormatFloat(form.Probability, 'f', 2, 64),
			form.Advice,
			form.Comment,
		}
		if err := writer.Write(record); err != nil {
			return FormShow{}, err
		}
	}
	logger.Infof("文件已保存到: %s\n", filePath)
	resp.Url = "/csv/" + filename
	return resp, nil
}

func (s *Form) FormDelete(formId int64) (err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	result := tx.Delete(&model.Form{}, "id = ?", formId)
	if result.Error != nil {
		tx.Rollback()
		return common.ErrNew(errors.New("删除委托失败"), common.SysErr)
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return common.ErrNew(errors.New("病历单号不存在"), common.ParamErr)
	}
	if err := tx.Commit().Error; err != nil {
		return common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}
	return nil
}
