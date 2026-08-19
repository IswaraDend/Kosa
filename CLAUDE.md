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
`backend-go/.env` (lihat `.env.example`): `DB_DSN`, `JWT_SECRET`, `CORS_ORIGINS`.
Kalau `JWT_SECRET` kosong, server memakai kunci acak per-boot dan memperingatkan —
sesi akan hangus tiap restart, jadi selalu set di deployment.

Cek sebelum menyerahkan perubahan:

```bash
cd backend-go && go build ./... && go vet ./... && gofmt -l .
```

```bash
cd frontend && npx tsc -b && npx eslint src
```

Belum ada test sama sekali di repo ini.

## Dua sumbu otorisasi

Ini konsep terpenting; keduanya berdiri sendiri.

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

| Peran | Dijaga oleh | Ikut permission | Ikut module |
|---|---|---|---|
| Super Admin | `RequireSuperAdmin` | tidak | tidak (sengaja) |
| Admin | `RequireProjectRole("admin")` | tidak | ya |
| Member | `RequirePermission` + `RequireModule` | ya | ya |

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

## Invarian domain

Dua alur stok terpisah: `Item` (bahan baku) lewat `Transaction` in/out/transfer →
`Stock`; `Product` (barang jadi) lewat `Production` (mengonsumsi `ProductRecipe`/BOM)
→ `ProductStock` → `Invoice`.

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

## Tema

Nama "Kosa" (कोष) berarti perbendaharaan sekaligus selubung. Kedelapan modul
dikelompokkan ke empat lapisan Pañca Kośa di sidebar (`LAYERS` di `lib/modules.ts`);
lapisan kelima, Ānandamaya, sengaja tanpa modul karena Ringkasan tidak pernah
di-gate. Palet tinta-kuningan, font Rozha One + IBM Plex, semuanya self-hosted.

## Bahasa

UI, pesan error API, dan komunikasi dengan pengguna memakai bahasa Indonesia.
Komentar dan identifier dalam kode memakai bahasa Inggris.
