package database

import (
	"backend-go/models"
	"log"
	"os"
	"strings"
	"time"

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
		dsn = "host=localhost user=postgres password=password123 dbname=kosa port=5432 sslmode=disable TimeZone=UTC"
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
		log.Println("WARNING: Failed to connect to database. Pastikan PostgreSQL menyala dan database 'kosa' tersedia.", err)
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
		&models.Customer{},
		&models.Invoice{},
		&models.InvoiceItem{},
		&models.ProjectModule{},
	)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	// AutoMigrate adds products.updated_at as NULL for rows that existed before
	// the column did, which would publish an empty date on the price list.
	// Seeding it from created_at is the closest true answer: that is the last
	// time those rows were written.
	if err := DB.Exec("UPDATE products SET updated_at = created_at WHERE updated_at IS NULL").Error; err != nil {
		log.Println("WARNING: gagal mengisi products.updated_at:", err)
	}

	SeedData()
}

func SeedData() {
	var count int64
	DB.Model(&models.User{}).Count(&count)

	if count == 0 {
		seedFirstSuperAdmin()
	}

	seedPermissions()
}

// seedFirstSuperAdmin creates the one account needed to log in to an empty
// database.
//
// The credentials come from SEED_SUPERADMIN_EMAIL / SEED_SUPERADMIN_PASSWORD.
// They used to be hardcoded, which meant every reader of this repository knew
// the super-admin login of any deployment seeded from it — the fallback below
// still exists so local development works out of the box, but it shouts about
// it, and it is the only path that also creates the throwaway admin fixture.
func seedFirstSuperAdmin() {
	email := strings.TrimSpace(os.Getenv("SEED_SUPERADMIN_EMAIL"))
	password := os.Getenv("SEED_SUPERADMIN_PASSWORD")
	isDefault := email == "" || password == ""

	if isDefault {
		email = "iswaradend@gmail.com"
		password = "password123"
		log.Println("WARNING: SEED_SUPERADMIN_EMAIL/SEED_SUPERADMIN_PASSWORD belum diset. " +
			"Memakai kredensial development yang tertulis di dalam source code — " +
			"JANGAN pakai ini di deployment yang bisa diakses publik.")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Gagal hash password super admin: ", err)
	}

	superAdmin := models.User{
		Name:         "Super Admin",
		Email:        email,
		Password:     string(hashed),
		IsSuperAdmin: true,
	}
	if err := DB.Create(&superAdmin).Error; err != nil {
		log.Fatal("Gagal membuat akun super admin awal: ", err)
	}

	if isDefault {
		// Development-only companion account; never created when real
		// credentials were supplied.
		DB.Create(&models.User{
			Name:         "Admin Cabang",
			Email:        "admin@perusahaan.com",
			Password:     string(hashed),
			IsSuperAdmin: false,
		})
	}

	log.Printf("Seeder: akun Super Admin '%s' dibuat.", email)
}

// seedPermissions is the single source of truth for the permission catalogue.
// Every Code here MUST correspond to a member route in routes.go guarded by
// middleware.RequirePermission with the same code — a permission that gates
// nothing is dead data that still shows up as a grantable checkbox.
//
// Permission.Module MUST be one of models.AllModules: the Admin permission
// screen only offers the permissions whose module is enabled for that project
// (see ListPermissionsForProject), so a permission carrying a module code that
// does not exist would be permanently invisible.
func seedPermissions() {
	defaults := []models.Permission{
		{Code: "warehouse.view", Name: "Lihat Gudang", Module: models.ModuleWarehouse},
		{Code: "warehouse.create", Name: "Tambah Gudang", Module: models.ModuleWarehouse},
		{Code: "warehouse.update", Name: "Ubah Gudang", Module: models.ModuleWarehouse},
		{Code: "warehouse.delete", Name: "Hapus Gudang", Module: models.ModuleWarehouse},

		{Code: "item.view", Name: "Lihat Item", Module: models.ModuleItem},
		{Code: "item.create", Name: "Tambah Item", Module: models.ModuleItem},
		{Code: "item.update", Name: "Ubah Item", Module: models.ModuleItem},
		{Code: "item.delete", Name: "Hapus Item", Module: models.ModuleItem},
		{Code: "item.import", Name: "Import Stok Item", Module: models.ModuleItem},

		{Code: "product.view", Name: "Lihat Produk", Module: models.ModuleProduct},
		{Code: "product.create", Name: "Tambah Produk", Module: models.ModuleProduct},
		{Code: "product.update", Name: "Ubah Produk & Resep", Module: models.ModuleProduct},
		{Code: "product.delete", Name: "Hapus Produk", Module: models.ModuleProduct},
		{Code: "product.import", Name: "Import Stok Produk", Module: models.ModuleProduct},

		{Code: "production.view", Name: "Lihat Produksi", Module: models.ModuleProduction},
		{Code: "production.create", Name: "Catat Produksi", Module: models.ModuleProduction},

		{Code: "transaction.view", Name: "Lihat Transaksi", Module: models.ModuleTransaction},
		{Code: "transaction.create", Name: "Buat Transaksi", Module: models.ModuleTransaction},

		{Code: "customer.view", Name: "Lihat Pelanggan", Module: models.ModuleCustomer},
		{Code: "customer.create", Name: "Tambah Pelanggan", Module: models.ModuleCustomer},
		{Code: "customer.update", Name: "Ubah Pelanggan", Module: models.ModuleCustomer},
		{Code: "customer.delete", Name: "Hapus Pelanggan", Module: models.ModuleCustomer},

		{Code: "invoice.view", Name: "Lihat Invoice", Module: models.ModuleInvoice},
		{Code: "invoice.create", Name: "Buat Invoice", Module: models.ModuleInvoice},
		{Code: "invoice.update", Name: "Ubah Status Invoice", Module: models.ModuleInvoice},

		{Code: "report.view", Name: "Lihat Laporan", Module: models.ModuleReport},
	}
	for _, p := range defaults {
		// Assign keeps Name/Module in sync on rows that already exist, so
		// renaming a permission here actually reaches an existing database
		// instead of silently applying only to fresh ones.
		if err := DB.Where(models.Permission{Code: p.Code}).
			Assign(models.Permission{Name: p.Name, Module: p.Module}).
			FirstOrCreate(&p).Error; err != nil {
			log.Println("WARNING: gagal seed permission", p.Code, err)
		}
	}
}
