# Kosa

Aplikasi manajemen stok, gudang, dan penjualan multi-tenant berbasis **project**.
Monorepo tiga layanan.

> Catatan: `../CLAUDE.md` dan `../AGENTS.md` di direktori induk mendeskripsikan
> proyek Figma Make yang berbeda dan **tidak berlaku di sini**. File ini yang
> berlaku untuk Kosa.

## Layanan

| Folder | Stack | Port | Status |
|---|---|---|---|
| `backend-go/` | Gin + GORM + PostgreSQL, JWT | 8080 | Inti sistem — seluruh logika bisnis |
| `frontend/` | React 18 + Vite + React Router 6, CSS vanilla | 5173 | Tiga dashboard: super-admin, admin, member |
| `backend-python/` | FastAPI | 8081 | **Stub** — baru `/ping`, belum dipakai |

Semua laporan dikerjakan Go (`handlers/report_handler.go`), bukan Python.

## Menjalankan

```bash
cd backend-go && go run main.go      # butuh PostgreSQL, database `kosa`
cd frontend && npm install && npm run dev
```

Tabel dibuat otomatis lewat AutoMigrate, lalu di-seed. Konfigurasi via
`backend-go/.env` (lihat `.env.example`): `DB_DSN`, `JWT_SECRET`, `CORS_ORIGINS`,
`API_KEY_ENC_KEY`.

Kalau `JWT_SECRET` kosong, server memakai kunci acak per-boot dan memperingatkan —
sesi akan hangus tiap restart, jadi selalu set di deployment.

`API_KEY_ENC_KEY` (32 byte hex) **wajib**, dan kalau kosong server sengaja
menolak boot — berbeda dari `JWT_SECRET`. Kunci acak aman untuk menandatangani
token, tapi di sini artinya semua secret key project yang sudah tersimpan tidak
bisa didekripsi lagi, dan fitur "lihat secret key" akan diam-diam mengembalikan
sampah. Buat sekali (`openssl rand -hex 32`) lalu simpan permanen.

AutoMigrate tidak pernah menghapus kolom, jadi perubahan skema yang sifatnya
membuang kolom ditulis manual di `migrateDropWarehouseColumns`
(`database/database.go`) dan dijalankan **sebelum** AutoMigrate. Semua langkah
di sana dijaga pengecekan `information_schema`, sehingga aman dijalankan ulang
dan tidak melakukan apa-apa di database baru.

Cek sebelum menyerahkan perubahan:

```bash
cd backend-go && go build ./... && go vet ./... && gofmt -l .
```

```bash
cd frontend && npx tsc -b && npx eslint src
```

Belum ada test sama sekali di repo ini.

## Tiga sumbu otorisasi

Ini konsep terpenting; ketiganya berdiri sendiri.

**Module** — batas komersial, melekat pada *project*. Ada 8 modul yang bisa
dinyalakan Super Admin per project (`models/module.go`, tabel `project_modules`):
`warehouse, item, product, production, transaction, customer, invoice, report`.
`middleware.RequireModule` menolak 403 kalau modulnya mati, tanpa peduli peran
pemanggil. Inilah yang mewujudkan "project hanya membayar fitur X".

Beberapa modul menarik dependensinya otomatis (`ExpandModules`), dan mematikan
modul ikut mematikan yang bergantung padanya (`CollapseModules`). `product`
sengaja **tidak** punya dependensi — sendirian ia hanya katalog barang jadi.

**Permission** — batas operasional, melekat pada *orang*, dan **hanya untuk
member**. `middleware.RequirePermission` mencari `roles.name = "member"` secara
harfiah, jadi grant pada role admin tidak pernah dibaca apa pun.

Ringkasnya:

**API key** — batas integrasi, melekat pada *project*, bukan pada orang. Satu
`ProjectAPIKey` ditukar jadi access token berumur 1 jam di `POST /api/v1/token`,
dan token itulah yang dipakai memanggil `/api/v1/*`. Pemanggilnya tidak punya
peran sama sekali dan tidak bisa menyentuh dashboard: token `/api/v1`
ditandatangani `config.APITokenKey` (turunan HMAC dari `JWTKey`), bukan `JWTKey`
itu sendiri, jadi kedua jenis token tidak saling berlaku. `RequireModule` tetap
dipasang di grup ini — secret key tidak boleh jadi jalan memutari batas modul.

| Peran | Dijaga oleh | Ikut permission | Ikut module |
|---|---|---|---|
| Super Admin | `RequireSuperAdmin` | tidak | tidak (sengaja) |
| Admin | `RequireProjectRole("admin")` | tidak | ya |
| Member | `RequirePermission` + `RequireModule` | ya | ya |
| Pemanggil API | `APITokenAuth` | tidak (bukan orang) | ya |

Katalog permission (`seedPermissions` di `database/database.go`) adalah satu-satunya
sumber kebenaran — 26 entri, satu per aksi yang bisa dijangkau member di kedelapan
modul. Setiap `Code` di sana **wajib** dipakai oleh rute member di
`routes/routes.go`, dan `Module`-nya wajib salah satu dari `models.AllModules` —
kalau tidak, permission itu tidak akan pernah muncul di layar mana pun.

Di frontend, `makeCan(permissions)` (`lib/permissions.ts`) memakai kode yang sama
persis untuk memutuskan tombol mana yang dirender. `undefined` berarti tanpa
gating per-aksi, yang benar untuk Super Admin dan Admin.

## Pola kode

**Scoping project.** Rute Super Admin bersifat global dengan `?project_id=`;
rute Admin/Member memakai path `/{admin,member}/projects/:projectId/...` dan
mengambil project dari context yang sudah divalidasi middleware — **jangan
pernah** membaca project dari request body pada jalur ini. Handler biasanya
berpasangan: `ListItems` (global) dan `ListItemsForProject` (scoped), keduanya
memanggil core function yang sama.

**Halaman frontend juga berpasangan.** Halaman di `pages/super-admin/` adalah
komponen induk yang dapat diparameterisasi; `pages/admin/` dan `pages/member/`
kebanyakan hanya pembungkus ~13 baris yang mengoper `apiBasePrefix`,
`scopeMode: 'query' | 'path'`, dan flag kemampuan. **Mengubah halaman
super-admin berarti mengubah halaman admin dan member juga.**

**Paginasi.** Endpoint list menerima `?page=` dan `?per_page=` dan membalas
amplop `{data, total, page, per_page, total_pages}`, di mana `total` adalah
jumlah baris yang cocok — bukan jumlah baris di `data`. Permintaan tanpa kedua
parameter itu dijawab utuh tanpa paginasi, sehingga pemanggil lama tetap jalan.
Helper-nya di `handlers/pagination.go`; sisi React memakai `usePagination` +
komponen `Pagination`. Pencarian ikut pindah ke server (`?q=`) di halaman yang
punya kotak cari — memfilter di klien hanya akan menyaring 25 baris yang sedang
tampil. Laporan yang berupa `GROUP BY` (tren transaksi, ringkasan penjualan,
produk terlaris) sengaja tidak dipaginasi.

**`lib/modules.ts` adalah cermin `models/module.go`** — daftar modul dan graf
dependensi ada di dua tempat dan harus diubah bersamaan.

**`ACTIVE_MODULES` (`lib/modules.ts`) adalah ruang lingkup build, bukan
konfigurasi project.** Saat ini isinya hanya `product, customer,
invoice` — aplikasi sengaja dipersempit ke alur Produk → Invoice. Modul di
luar daftar itu tidak dirender di sidebar mana pun, tidak punya rute di
`App.tsx`, dan tidak muncul sebagai centang di Kelola Project. Backend-nya
**utuh**: handler, rute Go, dan baris `project_modules` yang sudah ada tidak
disentuh, jadi mengembalikan sebuah fitur cukup dengan menambah kodenya ke
`ACTIVE_MODULES`. Karena itu, dua gerbang berlaku di frontend — `isActiveModule`
(build) **dan** modul yang diaktifkan project — dan `groupByLayer` menerapkan
keduanya. Project lama yang masih menyimpan modul di luar lingkup bisa
dimatikan lewat band "Di Luar Ruang Lingkup" di Kelola Project; band itu hanya
bisa mematikan, karena mematikan modul memicu `pruneMemberPermissions` yang
menghapus grant member secara permanen.

Stok produk bisa masuk tanpa Produksi lewat dua pintu yang memakai urutan
tulis yang sama (`applyProductCostLayer` lalu `adjustProductStock`): import
Excel, dan entri manual `POST .../products/:id/stock` (`adjustProductStockCore`).
Keduanya dijaga satu permission, `product.import` — mekanismenya berbeda,
aksinya sama.

**Gudang sudah dilepas dari jalur barang jadi.** `Invoice` tidak lagi punya
`WarehouseID`, dan `ProductStock` sekarang satu baris per `(project, produk)`,
bukan per gudang. Tabel `warehouses` **tetap ada** karena `Stock`,
`Transaction`, dan `Production` masih memakainya — menghapus model `Warehouse`
akan membuat seluruh rantai Item yang sedang disembunyikan itu gagal compile.
Jadi `warehouse` masih kode modul yang sah dan masih jadi dependensi `item`,
`transaction`, dan `production`; ia hanya bukan dependensi `invoice` lagi.

**Kategori produk** (`ProductCategory`) adalah bagian dari modul `product`,
bukan modul kesembilan, dan memakai ulang permission `product.*` satu lawan satu
(`view/create/update/delete`) — jadi katalog permission tetap 26 entri. Alasannya
sederhana: kategori hanya ada untuk dipasang ke produk, jadi grant yang
mengizinkan mengelola kategori tapi tidak produk tidak berarti apa-apa.
`Product.CategoryID` nullable, dan `nil` berarti "Tanpa kategori" — bukan nilai
yang masih harus diisi. Kategori yang masih dipakai produk tidak bisa dihapus
(409), bukan di-null-kan diam-diam.

## Invarian domain

Dua alur stok terpisah: `Item` (bahan baku) lewat `Transaction` in/out/transfer →
`Stock` (per gudang); `Product` (barang jadi) lewat `Production`, import Excel,
atau entri manual → `ProductStock` (per project) → `Invoice`.

- `AverageCost` dihitung rata-rata tertimbang dan **wajib diperbarui sebelum**
  kuantitas ditulis ke stok — rumusnya butuh qty pra-transaksi.
  Lihat `applyItemCostLayer` / `applyProductCostLayer`.
- `Invoice.TotalHPP` dan `InvoiceItem.UnitCOGS` adalah snapshot saat penjualan,
  supaya margin historis tidak bergeser saat `AverageCost` berubah.
- Membatalkan invoice mengembalikan `ProductStock`; invoice yang sudah
  `cancelled` tidak bisa diubah lagi, sehingga restock tidak pernah ganda.
- Semua pembaruan saldo berjalan di dalam DB transaction dengan
  `SELECT ... FOR UPDATE` (`clause.Locking`). Pertahankan itu saat menambah
  jalur baru yang menyentuh stok.
- `createTransactionTx` menerima `*gorm.DB` dari luar supaya Produksi dan Import
  Excel bisa menggabungkannya secara atomik ke transaksi mereka sendiri.

## API eksternal (`/api/v1`)

Grup rute tersendiri di `routes/routes.go`, sama sekali tidak lewat
`AuthMiddleware`. Alurnya dua langkah:

1. `POST /api/v1/token` dengan header `X-API-Key: kosa_sk_...` (atau
   `Authorization: Bearer kosa_sk_...`) → `{access_token, token_type,
   expires_in, project}`.
2. `GET /api/v1/products`, `/api/v1/products/:id`, `/api/v1/categories` dengan
   `Authorization: Bearer <access_token>`.

Yang perlu dijaga saat menambah endpoint di sini:

- **Selalu lewat DTO**, jangan pernah serialisasi model mentah. Lihat
  `handlers/api_v1_handler.go`. Kalau `models.Product` dikirim langsung, kolom
  apa pun yang ditambahkan nanti otomatis bocor ke semua pihak ketiga tanpa ada
  jejaknya di diff. Tidak ada `AverageCost`, HPP, atau margin di seluruh file itu.
- `projectID` **selalu** dari context yang diisi `APITokenAuth`, tidak pernah
  dari query string. Itu yang membuat isolasi antar-tenant jadi sifat
  middleware-nya, bukan sesuatu yang harus diingat tiap handler.
- Produk milik project lain dijawab **404**, bukan 403 — 403 membocorkan bahwa
  ID itu ada.
- Status key dan status project dicek ulang **tiap request**, bukan hanya saat
  menukar token, supaya key yang dicabut langsung mati dan tidak menunggu
  tokennya kadaluarsa.
- CORS sengaja **tidak** memasukkan `X-API-Key` ke `AllowHeaders`. API ini
  dirancang server-to-server; kalau dipanggil langsung dari browser, artinya
  secret key ada di dalam bundle JavaScript pihak lain.

Secret key disimpan dua kali dengan tujuan berbeda: `KeyHash` (SHA-256,
dibandingkan constant-time) untuk verifikasi tiap request, dan `KeyCipher`
(AES-256-GCM) hanya supaya Super Admin bisa membacanya kembali. Itu trade-off
yang disengaja — siapa pun yang memegang database **dan** `API_KEY_ENC_KEY` bisa
membaca semua kunci pelanggan — karena itu setiap `create`, `reveal`, dan
`revoke` ditulis ke `APIKeyAudit`, dan reveal yang gagal menulis audit ditolak.
Admin project boleh membuat dan mencabut kunci project-nya sendiri, tapi tidak
boleh `reveal`. Mencabut bersifat soft: barisnya tetap ada supaya jejak audit
masih punya rujukan.

## Tema

Nama "Kosa" (कोष) berarti perbendaharaan sekaligus selubung. Kedelapan modul
dikelompokkan ke empat lapisan Pañca Kośa di sidebar (`LAYERS` di `lib/modules.ts`);
lapisan kelima, Ānandamaya, sengaja tanpa modul karena Ringkasan tidak pernah
di-gate. Palet tinta-kuningan, font Rozha One + IBM Plex, semuanya self-hosted.

## Bahasa

UI, pesan error API, dan komunikasi dengan pengguna memakai bahasa Indonesia.
Komentar dan identifier dalam kode memakai bahasa Inggris.
