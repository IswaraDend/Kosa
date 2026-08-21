# Kosa — Task List Blocker

Dibuat 2026-08-21, cabang `development`.
Tiga blocker: **domain/deploy**, **secret key per project**, **API produk via secret key**.

Urutan kerja yang disarankan: **B2 → B3 → B1**. Secret key dan endpoint API bisa
dikerjakan dan diuji di `localhost` tanpa menunggu domain; domain hanya menahan
serah terima ke pihak ketiga, bukan pengembangannya.

---

## B1 — Belum punya domain

Yang sebenarnya terhambat cuma satu: pihak ketiga butuh URL stabil ber-TLS.
Pengembangan tidak perlu menunggu ini.

### Jalan pintas supaya tidak menunggu (kerjakan dulu)

- [ ] **B1.1** Sediakan URL sementara ber-HTTPS untuk integrasi awal — pilih salah satu:
      Cloudflare Tunnel (`cloudflared`, gratis, dapat subdomain `*.trycloudflare.com`),
      atau `sslip.io`/`nip.io` (`<ip>.sslip.io`) + Let's Encrypt.
      Tujuan: pihak ketiga bisa mulai coding sekarang, tinggal ganti base URL nanti.
- [ ] **B1.2** Pastikan base URL jadi variabel konfigurasi di sisi pihak ketiga, bukan hardcode.

### Setelah domain dibeli

- [ ] **B1.3** Beli domain + tentukan skema subdomain. Satu host sudah cukup:
      `docker-compose.yml` menyajikan SPA dan API di origin yang sama (nginx, `VITE_API_URL=/api`).
      Contoh: `app.<domain>` untuk dashboard + API.
- [ ] **B1.4** A record → IP VPS. Cek propagasi (`dig app.<domain>`).
- [ ] **B1.5** Reverse proxy + TLS di host (nginx/Caddy + Let's Encrypt, auto-renew).
- [ ] **B1.6** Set `PUBLIC_PORT=127.0.0.1:8080` di `.env`.
      **Penting:** Docker menulis aturan iptables sendiri dan melewati ufw — bentuk
      `8080:80` membuka dashboard + API ke internet walau `ufw status` terlihat menutup.
      Catatan ini sudah ada di `docker-compose.yml`, jangan sampai terlewat saat deploy.
- [ ] **B1.7** Set `PUBLIC_ORIGIN=https://app.<domain>`.
      `config.Load()` (`backend-go/config/config.go`) otomatis memasukkannya ke allow-list
      CORS, jadi `CORS_ORIGINS` cukup diisi pihak luar saja (storefront yang membaca
      `/public/projects/:code/prices` dari browser pengunjungnya).
- [ ] **B1.8** Set `TRUSTED_PROXIES` ke subnet reverse proxy — kalau kosong padahal ada proxy
      di depan, `X-Forwarded-For` tidak dipercaya dan rate limit (B3.7) salah menghitung IP.
- [ ] **B1.9** Set `JWT_SECRET` (`openssl rand -hex 32`). Kalau kosong, server memakai kunci
      acak per-boot dan semua sesi login hangus tiap restart.
- [ ] **B1.10** `docker compose build web` setelah `VITE_API_URL` berubah — nilai itu
      dipanggang ke dalam bundle saat build, restart saja tidak cukup.
- [ ] **B1.11** Perbarui `backend-go/.env.example` + `README.md` dengan variabel baru dari B2/B3.

---

## B2 — Secret key per project

**Keputusan desain yang harus diambil dulu (blocking untuk B2.2):**

> Permintaan "Super Admin bisa melihat semua secret key" berarti kunci **tidak bisa**
> disimpan sebagai hash saja. Dua opsi:
>
> - **Opsi A (paling aman):** simpan hash saja. Plaintext hanya muncul **sekali** saat
>   dibuat. Super Admin melihat daftar kunci (label, prefix, project, terakhir dipakai,
>   status) tapi tidak bisa membaca ulang nilainya — kalau hilang, rotate.
> - **Opsi B (sesuai permintaan harfiah):** simpan terenkripsi AES-256-GCM dengan
>   `API_KEY_ENC_KEY` dari env; hanya Super Admin yang boleh `reveal`.
>   Konsekuensi: siapa pun yang menguasai DB **dan** env bisa membaca semua kunci
>   pelanggan, dan setiap aksi reveal wajib punya audit trail.
>
> Task di bawah ditulis untuk **Opsi B**, dengan hash tetap disimpan untuk verifikasi
> cepat. Kalau nanti pilih A, cukup buang kolom cipher + endpoint reveal.

### Backend — model & penyimpanan

- [ ] **B2.1** Ambil keputusan Opsi A vs B di atas, catat alasannya di PR.
- [ ] **B2.2** Model baru `backend-go/models/project_api_key.go`:
      `ID, ProjectID, Name, KeyPrefix, KeyHash, KeyCipher, Status, LastUsedAt, ExpiresAt,
      CreatedBy, CreatedAt, RevokedAt`.
      `KeyHash` dan `KeyCipher` wajib `json:"-"` supaya tidak bocor lewat serialisasi tak sengaja.
      Index unik pada `KeyPrefix`, index biasa pada `ProjectID`.
- [ ] **B2.3** Daftarkan ke AutoMigrate di `backend-go/database/database.go`.
- [ ] **B2.4** Generator kunci: `kosa_sk_` + 32 byte `crypto/rand` (base62/base64url).
      `KeyPrefix` = 8 karakter pertama setelah prefix, disimpan terpisah supaya lookup
      lewat index — bukan scan seluruh tabel lalu bandingkan hash satu per satu.
- [ ] **B2.5** Enkripsi AES-256-GCM di `backend-go/config` (baca `API_KEY_ENC_KEY`, 32 byte hex).
      Server harus **gagal boot** kalau var ini kosong padahal Opsi B dipilih — jangan pakai
      fallback acak seperti `JWT_SECRET`, karena di sini kunci acak berarti data lama tak terbaca.
- [ ] **B2.6** Tabel audit `api_key_audits`: siapa, kunci mana, aksi `create|reveal|revoke`, kapan, IP.

### Backend — handler & rute

- [ ] **B2.7** `backend-go/handlers/api_key_handler.go`, core function dipakai bersama —
      ikuti pola `ListItems` / `ListItemsForProject`.
- [ ] **B2.8** Rute Super Admin di `routes/routes.go` (global, `?project_id=` opsional):
      - `GET    /super-admin/api-keys` — semua project, dipaginasi (`handlers/pagination.go`)
      - `POST   /super-admin/projects/:id/api-keys`
      - `POST   /super-admin/api-keys/:keyId/reveal` — hanya Super Admin, tulis audit
      - `DELETE /super-admin/api-keys/:keyId` — revoke (soft; jangan hard delete, jejak audit hilang)
- [ ] **B2.9** **Keputusan:** apakah Admin project boleh mengelola kunci project-nya sendiri?
      Rekomendasi: boleh **create + revoke** lewat `/admin/projects/:projectId/api-keys`,
      plaintext muncul **sekali** di respons create, tapi **tidak** boleh `reveal` kunci lama.
      Member: tidak ada akses sama sekali. Kalau tidak setuju, hapus grup admin ini seluruhnya.
- [ ] **B2.10** Pada jalur `/admin/...`, project **selalu** diambil dari context hasil
      `RequireProjectRole("admin")` — jangan pernah dari request body.
- [ ] **B2.11** Uji negatif: Admin project A tidak bisa membuat/melihat/mencabut kunci project B (403).

### Frontend

- [ ] **B2.12** `frontend/src/pages/super-admin/ApiKeyPage.tsx` — komponen induk
      terparameterisasi (`apiBasePrefix`, `scopeMode: 'query' | 'path'`, flag `canReveal`).
- [ ] **B2.13** Wrapper `pages/admin/` (~13 baris) kalau B2.9 diputuskan "boleh".
      Tidak ada wrapper member.
- [ ] **B2.14** UI: tampilkan plaintext **sekali** setelah create dengan tombol salin dan
      peringatan yang jelas. Selebihnya tampilkan bentuk tersamar `kosa_sk_ab12••••`.
- [ ] **B2.15** Pencarian dan paginasi di server (`?q=`, `?page=`, `?per_page=`) memakai
      `usePagination` + komponen `Pagination` — memfilter di klien hanya menyaring 25 baris
      yang sedang tampil.
- [ ] **B2.16** Entri sidebar. Halaman ini **tidak** di-gate modul (sama seperti kelola member
      dan permission) — ia infrastruktur project, bukan salah satu dari 8 modul.

---

## B3 — API `get product` dengan secret key

- [ ] **B3.1** `backend-go/middleware/api_key.go` — `APIKeyAuth()`:
      baca `X-API-Key` (fallback `Authorization: Bearer kosa_sk_...`), pecah prefix,
      lookup lewat prefix, verifikasi hash **constant-time** (`subtle.ConstantTimeCompare`).
      Tolak kalau dicabut / kedaluwarsa / project non-aktif.
      Set `projectID` ke context dengan key yang sama seperti middleware project yang sudah
      ada, supaya handler scoped bisa dipakai ulang tanpa perubahan.
- [ ] **B3.2** Update `LastUsedAt` dengan throttle (maksimal 1 tulis / 5 menit per kunci) —
      jangan `UPDATE` di setiap request, itu menulis ke DB pada jalur baca terpanas.
- [ ] **B3.3** Grup rute **baru dan terpisah** `/api/v1` di `routes/routes.go`. Jangan ditempel
      ke grup `/super-admin`, `/admin`, atau `/member` yang sudah memakai `AuthMiddleware()` JWT.
- [ ] **B3.4** Pasang `middleware.RequireModule(models.ModuleProduct)` di grup ini — kunci API
      tidak boleh jadi celah yang melewati batas komersial modul.
- [ ] **B3.5** `GET /api/v1/products` → `handlers.APIListProducts`, memanggil core function
      yang **sama** dengan `ListProductsForProject`, di-scope ke project dari context (B3.1).
      Dukung `?page=`, `?per_page=`, `?q=`; balas amplop
      `{data, total, page, per_page, total_pages}`.
- [ ] **B3.6** Bentuk respons: `id, sku, name, unit, default_price, created_at`.
      **Jangan** kirim `AverageCost` / HPP / margin — ikuti preseden `PublicPriceRow` di
      `handlers/public_handler.go`. Buat DTO tersendiri, jangan serialisasi `models.Product`
      mentah, supaya kolom baru di masa depan tidak ikut bocor otomatis.
- [ ] **B3.7** Rate limit per kunci (mis. 60 req/menit), balas `429` + header `Retry-After`.
- [ ] **B3.8** Bentuk error konsisten dan tidak informatif berlebih:
      `401` kunci hilang/salah/dicabut (jangan bedakan mana yang mana),
      `403` modul `product` mati, `429` rate limit. Pesan dalam bahasa Indonesia.
- [ ] **B3.9** CORS: `/api/v1` dirancang server-to-server. Kalau ada pemanggil dari browser,
      `X-API-Key` harus masuk `AllowHeaders` — tapi itu berarti kunci rahasia ada di dalam
      bundle JavaScript pihak ketiga. Default: **jangan** ditambahkan.
- [ ] **B3.10** Opsional: `GET /api/v1/products/:id`. Balas `404` kalau produk milik project
      lain, bukan `403` — `403` membocorkan bahwa ID itu ada.
- [ ] **B3.11** **Uji isolasi tenant:** kunci project A hanya bisa membaca produk project A.
      Ini uji paling penting di seluruh daftar ini.
- [ ] **B3.12** Dokumentasi pihak ketiga di `docs/api-v1.md`: base URL, cara kirim header,
      contoh `curl`, contoh respons, tabel kode error, kebijakan rate limit dan rotasi kunci.

---

## Lintas-blocker

- [ ] **X.1** `cd backend-go && go build ./... && go vet ./... && gofmt -l .`
- [ ] **X.2** `cd frontend && npx tsc -b && npx eslint src`
- [ ] **X.3** Perbarui `CLAUDE.md`: sekarang ada **sumbu otorisasi ketiga** di samping Module
      dan Permission, yaitu **API key** — melekat pada project, bukan pada orang.
      Tabel peran di sana perlu baris baru.
- [ ] **X.4** Belum ada test sama sekali di repo ini. Minimal tambahkan test Go untuk verifikasi
      kunci (B3.1) dan isolasi tenant (B3.11) — dua tempat yang kalau salah, salahnya senyap.
- [ ] **X.5** Rencana rotasi: apa yang terjadi kalau kunci pihak ketiga bocor? Minimal revoke
      instan, dan izinkan lebih dari satu kunci aktif per project supaya rotasi tidak
      menyebabkan downtime di sisi mereka.
