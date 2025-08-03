package models

import "gorm.io/gorm"

type Post struct {
	gorm.Model
	Title     string
	Body      string
	AccountID uint
}

// type Tag struct {
// 	gorm.Model
// 	Name string

// }
