# Kosa 📈

Aplikasi manajemen proyek, gudang, dan stok barang skala modern (Full-stack Monorepo).

## 🚀 Teknologi yang Digunakan
Proyek ini mengadopsi arsitektur *Monorepo* yang terdiri dari tiga layanan utama:
- **Frontend**: React + Vite (TypeScript, Vanilla CSS Dark Theme)
- **Backend Core**: Golang (Gin Framework + GORM)
- **Backend Data/Report**: Python
- **Database**: PostgreSQL

---

## 🛠️ Persiapan Database (PostgreSQL)

Sebelum menjalankan aplikasi, pastikan Anda telah membuat *database* di PostgreSQL.

1. Buka PostgreSQL (bisa melalui psql CLI atau pgAdmin).
2. Buat *database* baru bernama `kosa`:
   ```sql
   CREATE DATABASE kosa;
   ```
3. Konfigurasi kredensial *database* ada di dalam kode Go (secara _default_ menggunakan user: `[USERNAME]` dan password: `[PASSWORD]`). Anda bisa menyesuaikannya di file `backend-go/database/database.go`.
4. *Tabel-tabel database akan otomatis dibuat (Auto Migrate) saat server Golang pertama kali dijalankan.*

---

## 🏃‍♂️ Cara Menjalankan Aplikasi

Anda harus membuka 3 terminal terpisah untuk menjalankan seluruh ekosistem aplikasi ini.

### 1. Menjalankan Backend API (Golang)
Backend ini mengurus sistem *Role-Based Access Control* (RBAC), autentikasi JWT, dan logika inti aplikasi.

```bash
cd backend-go
go mod tidy
go run main.go
```
*Aplikasi akan berjalan di `http://localhost:8080`*

*(Catatan: Saat dijalankan, aplikasi akan otomatis melakukan 'seeding' untuk akun Super Admin. Anda dapat login menggunakan email: `[EMAIL_ADDRESS]` dan password: `[PASSWORD]`)*

### 2. Menjalankan Frontend (React)
Frontend ini merupakan Antarmuka Pengguna (*User Interface*) dengan tema gelap (*dark mode*) yang elegan.

```bash
cd frontend
npm install
npm run dev
```
*Aplikasi akan berjalan di `http://localhost:5173`*

### 3. Menjalankan Backend Service (Python)
Layanan Python digunakan untuk mendukung sistem *reporting* atau analisis data tambahan.

```bash
cd backend-python
# Buat Virtual Environment (Sangat disarankan)
python -m venv venv

# Aktifkan Virtual Environment
# Di Windows:
venv\Scripts\activate
# Di Mac/Linux:
# source venv/bin/activate

# Jalankan script utama
python main.py
```

---

## 📚 Struktur Folder
```text
stock/
├── backend-go/        # Source code Golang (API, Models, DB Config)
├── backend-python/    # Source code Python (Services/Reporting)
├── frontend/          # Source code React (Vite, UI, Pages)
├── .gitignore         # Konfigurasi pengecualian sistem Git
└── README.md          # Dokumentasi proyek
```

## 🔐 Role-Based Access Control (RBAC)
Sistem ini membagi otorisasi menjadi 3 tingkat dasbor (berdasarkan tabel relasi GORM):
1. **Super Admin**: Mengelola seluruh *project*, pembuatan akun admin, dll.
2. **Admin**: Mengelola anggota (*member*) dan memberikan *permissions* (hak akses fitur).
3. **Member**: Menerima hak akses spesifik per modul.
