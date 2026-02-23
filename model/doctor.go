package model

type Doctor struct {
	Username string `gorm:"type:VARCHAR(128) NOT NULL;comment:名称" json:"username"`
	Password string `gorm:"type:VARCHAR(256) NOT NULL;comment:密码" json:"password"`
	Phone    string `gorm:"type:VARCHAR(128) NOT NULL;comment:手机号" json:"phone"`
	Level    int    `gorm:"type:INT NOT NULL DEFAULT 0;comment:医生级别" json:"level"`
	Form     []Form `gorm:"foreignKey:DoctorID" json:"forms"`
	BaseModel
}

func (Doctor) TableName() string {
	return "doctors"
}
