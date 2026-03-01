package database

import (
    "errors"
    "gorm.io/gorm"
)

func Add[T any](db *gorm.DB, data *T) error {
    if db == nil {
        return errors.New("db connection is nil")
    }
    
    result := db.Create(data)
    
    return result.Error
}