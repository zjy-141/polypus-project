package model

type Patient struct {
	PatientNumber string `gorm:"type:VARCHAR(256) NOT NULL;comment:门诊号;uniqueIndex" json:"patient_number"`
	Forms         []Form `gorm:"foreignKey:PatientNumber;references:PatientNumber" json:"forms"`
	BaseModel
}

func (Patient) TableName() string {
	return "patients"
}
