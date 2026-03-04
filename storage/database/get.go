package database

import (
    "errors"
    "time"
    "telescrap/structs"
    "gorm.io/gorm"
)

func Get[T any](db *gorm.DB, opts ...Option) ([]T, error) {
    if db == nil {
        return nil, errors.New("db connection is nil")
    }

    var objects []T

    query := db

    for _, opt := range opts {
        query = opt(query)
    }

    err := query.Find(&objects).Error
    return objects, err
}

type Option func(db *gorm.DB) *gorm.DB

func WithDate(from, to time.Time) Option {
    return func(db *gorm.DB) *gorm.DB {
        if !to.IsZero() {
            db = db.Where("created_at >= ?", to)
        }

        if !from.IsZero() {
            db = db.Where("created_at <= ?", from)
        }

        return db
    }
}

func WithID(id string) Option {
    return func(db *gorm.DB) *gorm.DB {
        if id != "" {
            db = db.Where("id = ?", id)
        }
        return db
    }
}

func WithTGID(tgid string) Option {
    return func(db *gorm.DB) *gorm.DB {
        if tgid != "" {
            db = db.Where("tgid = ?", tgid)
        }

        return db
    }
}

func WithGroupID(groupID string) Option {
    return func(db *gorm.DB) *gorm.DB {
        if groupID != "" {
            db = db.Where("group_id = ?", groupID)
        }

        return db
    }
}

func IsHere(id int, db *gorm.DB) bool {
    var count int64
    db.Model(&structs.User{}).Where("ID = ?", id).Count(&count)
    return count > 0
}