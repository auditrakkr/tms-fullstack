package models

import "gorm.io/gorm"


type Role struct {
	gorm.Model
	Name string `gorm:"type:varchar(255);not null;unique"`
	Description string `gorm:"type:text"`
	Users []User `gorm:"many2many:user_roles;"`
	Landlord bool `gorm:"default:true"`
	
	//Todo permissions relationship for each role
}