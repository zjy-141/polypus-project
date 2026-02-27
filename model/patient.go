package model

type Patient struct {
	ClinicNumber string `gorm:"type:VARCHAR(256) NOT NULL;comment:门诊号;uniqueIndex" json:"clinic_number"`
	Forms        []Form `gorm:"foreignKey:ClinicNumber;references:ClinicNumber" json:"forms"`
	BaseModel
}

func (Patient) TableName() string {
	return "patients"
}
