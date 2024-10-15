package repositories

import (
	"fmt"
	"gombot/pkg/configs"
	"gombot/pkg/entities"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

var isInitialized = false

func DbConnect() *gorm.DB {
	cfg := configs.LoadConfig(configs.ConfigPath)
	connStr := fmt.Sprintf("host=%s user=%s password=%s port=%d sslmode=disable TimeZone=Asia/Tehran",
		cfg.DatabaseConfig.Host, cfg.DatabaseConfig.Username, cfg.DatabaseConfig.Password, cfg.DatabaseConfig.Port)
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		panic("failed to connect database") //TODO: study about panic
	}

	if cfg.DatabaseConfig.NeedMigration && !isInitialized {
		isInitialized = true
		migrate(db)
	}

	return db
}

func migrate(db *gorm.DB) {
	err := db.AutoMigrate(&entities.Job{}, &entities.Application{},
		&entities.Approver{}, &entities.Requester{})
	if err != nil {
		log.Panicf("can not migrate entities. there is error %v", err) // TODO: log.Panicf() vs. Panic()
	}
}

func examples() {
	db := DbConnect()

	// Migrate the schema

	// Create
	db.Create(&Product{Code: "D42", Price: 100})

	// Read
	var product Product
	db.First(&product, 1)                 // find product with integer primary key
	db.First(&product, "code = ?", "D42") // find product with code D42

	// Update - update product's price to 200
	db.Model(&product).Update("Price", 200)
	// Update - update multiple fields
	db.Model(&product).Updates(Product{Price: 200, Code: "F42"}) // non-zero fields
	db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

	// Delete - delete product
	db.Delete(&product, 1)

}
