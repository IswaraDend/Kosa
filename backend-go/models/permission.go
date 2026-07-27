package models

type Permission struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Module string `json:"module"`
}
