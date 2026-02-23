package model

type Form struct {
	DoctorID      int64    `gorm:"type:VARCHAR(128) NOT NULL;comment:医生ID" json:"doctor_id"`
	PatientNumber string   `gorm:"type:VARCHAR(256) NOT NULL;comment:门诊号" json:"patient_number"`
	FormTime      string   `gorm:"type:VARCHAR(256) NOT NULL;comment:表单时间" json:"form_time"`
	Age           int      `gorm:"type:INT NOT NULL DEFAULT 0;comment:年龄" json:"age"`
	LongDiameter  float64  `gorm:"type:FLOAT NOT NULL DEFAULT 0;comment:长径" json:"long_diameter"`
	ShortDiameter float64  `gorm:"type:FLOAT NOT NULL DEFAULT 0;comment:短径" json:"short_diameter"`
	BaseType      int      `gorm:"type:INT NOT NULL DEFAULT 0;comment:基底类型" json:"base_type"`
	PolypNumber   int      `gorm:"type:INT NOT NULL DEFAULT 0;comment:息肉数量" json:"polyp_number"`
	RiskLevel     string   `gorm:"type:VARCHAR(256) NOT NULL DEFAULT 0;comment:风险等级" json:"risk_level"`
	Probability   float64  `gorm:"type:FLOAT NOT NULL DEFAULT 0;comment:预测准确概率" json:"probability"`
	Advice        string   `gorm:"type:TEXT;comment:建议" json:"advice"`
	IsWorse       string   `gorm:"type:VARCHAR(128) NOT NULL DEFAULT 0;comment:是否恶化" json:"is_worse"`
	Comment       string   `gorm:"type:TEXT;comment:备注" json:"comment"`
	Doctor        *Doctor  `gorm:"foreignKey:DoctorID" json:"doctor"`
	Patient       *Patient `gorm:"foreignKey:PatientNumber;references:ClinicNumber" json:"patient"`
	BaseModel
}

func (Form) TableName() string {
	return "forms"
}
