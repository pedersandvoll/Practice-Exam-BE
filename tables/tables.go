package tables

import (
	"time"

	"gorm.io/gorm"
)

type Priority int

const (
	High Priority = iota
	Medium
	Low
)

type Users struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"size:100"`
	Email     string    `gorm:"uniqueIndex"`
	Password  string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type Customers struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"size:100;uniqueIndex"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type Complaints struct {
	ID          uint      `gorm:"primaryKey"`
	CustomerID  uint      `gorm:"not null"`
	Customer    Customers `gorm:"foreignKey:CustomerID"`
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	ModifiedAt  time.Time `gorm:"autoUpdateTime"`
	CreatedByID uint      `gorm:"not null"`
	CreatedBy   Users     `gorm:"foreignKey:CreatedByID"`
	Priority    Priority
	Comments    []Comments `gorm:"foreignKey:ComplaintID"`
}

type Comments struct {
	ID          uint      `gorm:"primaryKey"`
	Comment     string    `gorm:"type:text"`
	ComplaintID uint      `gorm:"not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	CreatedByID uint      `gorm:"not null"`
	CreatedBy   Users     `gorm:"foreignKey:CreatedByID"`
}

func RunMigrations(db *gorm.DB) {
	db.AutoMigrate(&Users{}, &Customers{}, &Complaints{}, &Comments{})
}
