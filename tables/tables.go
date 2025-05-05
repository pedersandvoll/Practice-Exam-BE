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
	ModifiedAt  time.Time `gorm:"autoCreateTime"`
	CreatedByID uint      `gorm:"not null"`
	CreatedBy   Users     `gorm:"foreignKey:CreatedByID"`
	Priority    Priority
}

func RunMigrations(db *gorm.DB) {
	db.AutoMigrate(&Users{}, &Customers{}, &Complaints{})
}
