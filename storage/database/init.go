package database

import(
	"gorm.io/gorm"
	"telescrap/structs"
	"log"
)

func Init(db *gorm.DB) error {


	err := db.AutoMigrate(&structs.Message{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
		return err
	}

	err = db.AutoMigrate(&structs.Group{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
		return err
	}

	err = db.AutoMigrate(&structs.User{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
		return err
	}	
	return nil
}