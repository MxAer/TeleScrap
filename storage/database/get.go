package database

import (
    "errors"
    "time"
    "telescrap/structs"
    "gorm.io/gorm"
)

func Get[T any](dateFrom time.Time, dateTo time.Time, db *gorm.DB) ([]T, error) {
    if db == nil {
        return nil, errors.New("db connection is nil")
    }

    var objects []T

    query := db

    if !dateFrom.IsZero() {
        query = query.Where("created_at >= ?", dateFrom)
    }
    if !dateTo.IsZero() {
        query = query.Where("created_at <= ?", dateTo)
    }

    err := query.Find(&objects).Error
    return objects, err
}

func IsHere(id int, db *gorm.DB) bool {
    var count int64 
    
    db.Model(&structs.User{}).Where("ID = ?", id).Count(&count)  
    
    return count > 0
}