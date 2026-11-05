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

	// Both run before AutoMigrate on purpose — see the comments on them.
	migrateDropWarehouseColumns()
	migrateBackfillProductCategoryCode()

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
		&models.ProductCategory{},
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
		&models.ProjectAPIKey{},
		&models.APIKeyAudit{},
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
		// Covers both doors onto ProductStock: the Excel import and the manual
		// entry on the product page. Assign() below means this rename reaches
		// databases that already hold the old label.
		{Code: "product.import", Name: "Atur Stok Produk (import & manual)", Module: models.ModuleProduct},

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

// columnExists answers from the catalog rather than from a failed ALTER, so a
// migration that has already run is skipped silently instead of logging an
// error on every boot.
func columnExists(table, column string) bool {
	var n int64
	err := DB.Raw(`
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = current_schema()
		  AND table_name = ? AND column_name = ?
	`, table, column).Scan(&n).Error
	if err != nil {
		log.Println("WARNING: gagal memeriksa kolom", table+"."+column, err)
		return false
	}
	return n > 0
}

// tableExists keeps a migration from touching a table AutoMigrate has not
// created yet, which is the normal state on a fresh database.
func tableExists(table string) bool {
	var n int64
	err := DB.Raw(`
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = current_schema() AND table_name = ?
	`, table).Scan(&n).Error
	if err != nil {
		log.Println("WARNING: gagal memeriksa tabel", table, err)
		return false
	}
	return n > 0
}

// migrateBackfillProductCategoryCode gives every existing category a code
// before AutoMigrate puts a unique index on (project_id, code).
//
// Without this, a database that already holds categories gets the new column
// filled with an empty string on every row, and the unique index then cannot be
// created at all: AutoMigrate fails and the server refuses to boot. Adding it
// here and filling it first means AutoMigrate only has to add the index to
// data that already satisfies it.
//
// The code is derived from the name, upper-cased and stripped to letters and
// digits, which usually produces something the owner would have typed anyway.
// Only the rows that would still collide get the id appended, so one awkward
// duplicate does not uglify every other code.
func migrateBackfillProductCategoryCode() {
	if !tableExists("product_categories") || columnExists("product_categories", "code") {
		return
	}

	// Added bare, with no constraint: AutoMigrate adds the unique index right
	// after, once the rows below can satisfy it.
	if err := DB.Exec("ALTER TABLE product_categories ADD COLUMN code text").Error; err != nil {
		log.Fatal("Migration failed: gagal menambah product_categories.code: ", err)
	}

	if err := DB.Exec(`
		UPDATE product_categories
		SET code = COALESCE(
			NULLIF(UPPER(SUBSTRING(REGEXP_REPLACE(name, '[^a-zA-Z0-9]', '', 'g') FROM 1 FOR 16)), ''),
			'KAT'
		)
	`).Error; err != nil {
		log.Fatal("Migration failed: gagal mengisi product_categories.code: ", err)
	}

	// Two names can reduce to the same code ("Kopi Susu" and "Kopisusu"), and
	// a name with no letters at all falls back to the shared 'KAT'.
	if err := DB.Exec(`
		UPDATE product_categories pc SET code = pc.code || '-' || pc.id
		WHERE EXISTS (
			SELECT 1 FROM product_categories other
			WHERE other.project_id = pc.project_id
			  AND other.code = pc.code
			  AND other.id <> pc.id
		)
	`).Error; err != nil {
		log.Fatal("Migration failed: gagal membuat kode kategori unik: ", err)
	}

	log.Println("Migrasi: kolom code ditambahkan ke product_categories dan diisi dari nama")
}

// migrateDropWarehouseColumns takes finished-goods stock and invoices off
// Warehouse.
//
// It has to run BEFORE AutoMigrate: the unique index on product_stocks is now
// (project_id, product_id), and a database still holding one row per warehouse
// would have duplicates under that key, so AutoMigrate could not create the
// index at all. Every step is guarded by the column still being there, which
// makes the whole thing a no-op on a fresh database and on every boot after
// the first.
//
// Warehouses themselves are deliberately left alone — Item stock, Transaction
// and Production still reference them.
func migrateDropWarehouseColumns() {
	if columnExists("product_stocks", "warehouse_id") {
		// Two warehouses holding the same product collapse into one balance.
		// Summing is the only answer that keeps the total on hand correct;
		// keeping one row and dropping the others would destroy real stock.
		if err := DB.Exec(`
			UPDATE product_stocks ps SET quantity = agg.total
			FROM (
				SELECT project_id, product_id, MIN(id) AS keep_id, SUM(quantity) AS total
				FROM product_stocks GROUP BY project_id, product_id
			) agg
			WHERE ps.id = agg.keep_id
		`).Error; err != nil {
			log.Fatal("Migration failed: gagal menggabungkan stok produk antar gudang: ", err)
		}
		if err := DB.Exec(`
			DELETE FROM product_stocks ps
			USING (
				SELECT project_id, product_id, MIN(id) AS keep_id
				FROM product_stocks GROUP BY project_id, product_id
			) agg
			WHERE ps.project_id = agg.project_id
			  AND ps.product_id = agg.product_id
			  AND ps.id <> agg.keep_id
		`).Error; err != nil {
			log.Fatal("Migration failed: gagal membersihkan stok produk ganda: ", err)
		}
		// Dropping the column takes idx_warehouse_product with it, since that
		// index is built partly on it.
		if err := DB.Exec("ALTER TABLE product_stocks DROP COLUMN warehouse_id").Error; err != nil {
			log.Fatal("Migration failed: gagal menghapus product_stocks.warehouse_id: ", err)
		}
		log.Println("Migrasi: stok produk digabung per project, kolom warehouse_id dihapus")
	}

	if columnExists("invoices", "warehouse_id") {
		if err := DB.Exec("ALTER TABLE invoices DROP COLUMN warehouse_id").Error; err != nil {
			log.Fatal("Migration failed: gagal menghapus invoices.warehouse_id: ", err)
		}
		log.Println("Migrasi: kolom invoices.warehouse_id dihapus")
	}
}
