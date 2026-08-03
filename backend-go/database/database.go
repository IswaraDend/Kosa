package database

import (
	"backend-go/models"
	"log"
	"time"

	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		// Fallback for development if .env is missing
		dsn = "host=localhost user=postgres password=password123 dbname=stockpulse port=5432 sslmode=disable TimeZone=UTC"
	}

	gormLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormLogger})
	if err != nil {
		log.Println("WARNING: Failed to connect to database. Pastikan PostgreSQL menyala dan database 'stockpulse' tersedia.", err)
		return
	}

	DB = db
	log.Println("Database PostgreSQL connected!")

	// Auto Migrate Table
	err = DB.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.Role{},
		&models.UserRole{},
		&models.Permission{},
		&models.MemberPermission{},
		&models.Warehouse{},
		&models.Item{},
		&models.Stock{},
		&models.Product{},
		&models.ProductRecipe{},
		&models.ProductStock{},
		&models.Transaction{},
		&models.TransactionItem{},
		&models.Production{},
	)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	SeedData()
}

func SeedData() {
	var count int64
	DB.Model(&models.User{}).Count(&count)

	if count == 0 {
		// Generate hashed password "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

		superAdmin := models.User{
			Name:         "Iswara Super Admin",
			Email:        "iswaradend@gmail.com",
			Password:     string(hashedPassword),
			IsSuperAdmin: true,
		}

		admin := models.User{
			Name:         "Admin Cabang",
			Email:        "admin@perusahaan.com",
			Password:     string(hashedPassword),
			IsSuperAdmin: false,
		}

		DB.Create(&superAdmin)
		DB.Create(&admin)
		log.Println("Seeder executed! Akun iswaradend@gmail.com (Super Admin) dan admin@perusahaan.com telah dibuat. Default password: password123")
	}

	seedPermissions()
}

func seedPermissions() {
	defaults := []models.Permission{
		{Code: "warehouse.view", Name: "Lihat Gudang", Module: "warehouse"},
		{Code: "item.view", Name: "Lihat Item", Module: "item"},
		{Code: "transaction.view", Name: "Lihat Transaksi", Module: "transaction"},
		{Code: "transaction.create", Name: "Buat Transaksi", Module: "transaction"},
		{Code: "report.view", Name: "Lihat Laporan", Module: "report"},
	}
	for _, p := range defaults {
		DB.Where(models.Permission{Code: p.Code}).FirstOrCreate(&p)
	}
}
