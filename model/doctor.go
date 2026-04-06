package model

type Doctor struct {
	Username string `gorm:"type:VARCHAR(128) NOT NULL;comment:名称" json:"username"`
	Password string `gorm:"type:VARCHAR(256) NOT NULL;comment:密码" json:"password"`
	Phone    string `gorm:"type:VARCHAR(128) NOT NULL;comment:手机号" json:"phone"`
	Level    int    `gorm:"type:INT NOT NULL DEFAULT 0;comment:医生级别,-1被删除,1医生,2超级管理员" json:"level"`
	Reset    int    `gorm:"type:INT NOT NULL DEFAULT 0;comment:是否要求重置密码,1要求,2不要" json:"reset"`
	Form     []Form `gorm:"foreignKey:DoctorID" json:"forms"`
	BaseModel
}

func (Doctor) TableName() string {
	return "doctors"
}
