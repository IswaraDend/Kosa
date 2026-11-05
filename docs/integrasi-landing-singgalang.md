# Integrasi Landing Page — Singgalang KSB

Panduan untuk pengembang landing page Singgalang KSB yang mengambil daftar
harga dari Kosa. Referensi lengkap seluruh endpoint ada di
[`api-v1.md`](api-v1.md); dokumen ini hanya bagian yang Anda perlukan.

## Yang Anda butuhkan

| | |
|---|---|
| Base URL | `https://api.kosa.co.id/v1` |
| Secret key | Diberikan terpisah, berawalan `kosa_sk_` |
| Project | Singgalang-ksb |

**Secret key tidak ditulis di dokumen ini dengan sengaja.** Dokumen bisa
berpindah tangan, kunci tidak boleh. Simpan di environment variable di sisi
server (`KOSA_API_KEY=...`), jangan di dalam kode yang ikut ter-build.

Alamat lama `https://app.kosa.co.id/api/v1/...` tetap berlaku dan menunjuk ke
API yang sama. Perhatikan prefiksnya berbeda: `/v1` di `api.kosa.co.id`,
`/api/v1` di `app.kosa.co.id`. Mencampurnya jadi `api.kosa.co.id/api/v1/...`
menjawab 404.

## Alur dua langkah

Secret key tidak dipakai langsung untuk membaca data. Ia ditukar dulu menjadi
access token berumur **1 jam**, dan token itulah yang dipakai memanggil
endpoint data.

### 1. Tukar kunci jadi token

```bash
curl -X POST https://api.kosa.co.id/v1/token \
  -H "X-API-Key: $KOSA_API_KEY"
```

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "project": { "id": 1, "code": "sksb", "name": "Singgalang-ksb" }
}
```

### 2. Ambil produk

```bash
curl https://api.kosa.co.id/v1/products?category=LOGAM-MULIA-ANTAM \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

## Kategori

Dua kategori, dipanggil lewat **kode**-nya, bukan id angka. Kode tetap sama
walau nama kategorinya diubah di dashboard, jadi tautan di sisi Anda tidak ikut
mati.

| Kode | Nama | Isi | Satuan |
|---|---|---|---|
| `LOGAM-MULIA-ANTAM` | Logam Mulia Antam | 12 produk | keping |
| `EMAS-KARAT` | Emas Karat | 23 produk | gram |

Kode tidak peka huruf besar/kecil: `?category=emas-karat` sama saja.

Tanpa parameter `category`, seluruh 35 produk dikembalikan.

Daftar kategori juga bisa diambil sendiri, berguna kalau nanti bertambah:

```bash
curl https://api.kosa.co.id/v1/categories \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

## Bentuk balasan

```json
{
  "data": [
    {
      "id": 24,
      "sku": "LM-ANTAM-0.5G",
      "name": "Logam Mulia ANTAM 0.5 gram",
      "unit": "keping",
      "default_price": 1360000,
      "category": {
        "id": 1,
        "code": "LOGAM-MULIA-ANTAM",
        "name": "Logam Mulia Antam"
      },
      "updated_at": "2026-11-05T15:13:18.951582Z"
    }
  ],
  "total": 12,
  "page": 1,
  "per_page": 25,
  "total_pages": 1
}
```

Catatan yang sering jadi salah paham:

- `default_price` adalah **angka rupiah polos**, bukan string dan tanpa
  pemisah ribuan. Formatnya urusan sisi Anda:
  `new Intl.NumberFormat('id-ID').format(1360000)` → `1.360.000`.
- `total` adalah jumlah baris yang cocok, **bukan** jumlah baris di `data`.
- `updated_at` adalah waktu produk itu terakhir diubah, dan dikirim dalam
  **UTC** — perhatikan akhiran `Z`, bukan `+07:00`. Konversi ke WIB sebelum
  ditampilkan, kalau tidak harga yang diperbarui jam 00:30 WIB akan tertulis
  tanggal kemarin. Pakai nilai paling baru di antara semua baris untuk menulis
  "Harga per \<tanggal\>" di halaman Anda — lebih jujur daripada menampilkan
  tanggal hari ini.
- `category` bernilai `null` untuk produk tanpa kategori. Saat ini tidak ada,
  tapi jangan berasumsi ia selalu ada.

## Query lain

| Query | Arti |
|---|---|
| `category` | Saring per kode kategori |
| `q` | Cari di nama atau SKU, tidak peka huruf besar/kecil |
| `page`, `per_page` | Paginasi. Tanpa keduanya, semua baris dikirim sekaligus |

Satu produk saja: `GET /v1/products/:id`, bentuknya sama dengan satu
elemen `data`.

## Urutan

API tidak menjamin urutan tertentu. Untuk tabel harga, urutkan sendiri di sisi
Anda — misalnya emas karat menurun dari 24K, dan logam mulia menaik dari 0.5
gram. Mengandalkan urutan bawaan akan membuat tabel Anda berubah susunan tanpa
sebab yang terlihat.

## Kode error

| Status | Arti | Tindakan |
|---|---|---|
| 401 saat tukar token | Secret key salah, dicabut, atau project nonaktif | Periksa key; jangan ulangi otomatis |
| 401 saat ambil data | Access token kedaluwarsa (lewat 1 jam) | Tukar token baru, ulangi sekali |
| 403 | Modul `product` dimatikan untuk project ini | Hubungi admin Kosa |
| 404 | Produk tidak ada, atau milik project lain | — |
| 429 | Lebih dari 60 permintaan/menit per kunci | Lihat caching di bawah |

Kode kategori yang tidak dikenal **bukan** error: jawabannya 200 dengan `data`
kosong. "Tidak ada produk di kategori itu" adalah jawaban yang sah.

## Caching dan batas pemakaian

Batasnya **60 permintaan per menit per kunci**. Itu banyak untuk pemakaian
normal, tapi hanya kalau Anda tidak memanggil API pada setiap kunjungan
pengunjung.

Yang disarankan:

- **Panggil dari server, bukan dari browser pengunjung.** Harga diambil di
  sisi server landing page, lalu disajikan sebagai HTML atau JSON statis.
- **Simpan access token dan pakai ulang** selama satu jam. Menukar token baru
  di setiap permintaan membakar jatah tanpa guna.
- **Cache daftar produknya**, 5–15 menit sudah cukup. Harga toko berubah
  harian, bukan per detik.
- Kalau memakai Next.js atau sejenisnya, `revalidate: 600` pada fetch-nya
  sudah menyelesaikan semua hal di atas sekaligus.

## Keamanan

API ini dirancang **server-to-server**. Header `X-API-Key` sengaja tidak masuk
daftar CORS, jadi memanggil langkah tukar token dari browser pengunjung tidak
akan berhasil — dan itu memang disengaja: kalau berhasil, artinya secret key
ada di dalam bundle JavaScript yang bisa dibaca siapa saja.

Kalau benar-benar perlu memanggil dari browser, yang boleh ada di sana hanya
`access_token` (umur 1 jam) yang diterbitkan server Anda sendiri — dan origin
landing page harus didaftarkan lebih dulu ke `CORS_ORIGINS` di sisi Kosa.

Kunci yang bocor bisa dicabut kapan saja dari dashboard Kosa (menu **Secret
Key**). Pencabutan berlaku seketika, tidak menunggu token yang sudah terbit
kedaluwarsa.

## Contoh lengkap (Node.js)

```js
const BASE = 'https://api.kosa.co.id/v1';
const KEY = process.env.KOSA_API_KEY;

let cached = { token: null, expiresAt: 0 };

async function getToken() {
  // Diperbarui 60 detik sebelum kedaluwarsa supaya tidak pernah ada
  // permintaan yang berangkat dengan token yang baru saja mati.
  if (cached.token && Date.now() < cached.expiresAt - 60_000) return cached.token;

  const res = await fetch(`${BASE}/token`, {
    method: 'POST',
    headers: { 'X-API-Key': KEY },
  });
  if (!res.ok) throw new Error(`Tukar token gagal: ${res.status}`);

  const data = await res.json();
  cached = { token: data.access_token, expiresAt: Date.now() + data.expires_in * 1000 };
  return cached.token;
}

export async function getProducts(categoryCode) {
  const token = await getToken();
  const url = new URL(`${BASE}/products`);
  if (categoryCode) url.searchParams.set('category', categoryCode);

  const res = await fetch(url, { headers: { Authorization: `Bearer ${token}` } });
  if (!res.ok) throw new Error(`Ambil produk gagal: ${res.status}`);

  const { data } = await res.json();
  return data;
}

// Pemakaian
const logamMulia = await getProducts('LOGAM-MULIA-ANTAM');
const emasKarat = await getProducts('EMAS-KARAT');
```

## Isi katalog saat ini

Per 5 November 2026. Angka di bawah hanya untuk mencocokkan saat pertama kali
menyambungkan — **jangan ditulis permanen di kode Anda**, karena harga toko
berubah harian dan sumber kebenarannya adalah API.

### `LOGAM-MULIA-ANTAM` — 12 produk, satuan keping

| SKU | Nama | Harga |
|---|---|---|
| `LM-ANTAM-0.5G` | Logam Mulia ANTAM 0.5 gram | 1.360.000 |
| `LM-ANTAM-1G` | Logam Mulia ANTAM 1 gram | 2.572.000 |
| `LM-ANTAM-2G` | Logam Mulia ANTAM 2 gram | 5.066.000 |
| `LM-ANTAM-3G` | Logam Mulia ANTAM 3 gram | 7.580.000 |
| `LM-ANTAM-5G` | Logam Mulia ANTAM 5 gram | 12.545.000 |
| `LM-ANTAM-10G` | Logam Mulia ANTAM 10 gram | 25.030.000 |
| `LM-ANTAM-25G` | Logam Mulia ANTAM 25 gram | 62.370.000 |
| `LM-ANTAM-50G` | Logam Mulia ANTAM 50 gram | 124.600.000 |
| `LM-ANTAM-100G` | Logam Mulia ANTAM 100 gram | 249.000.000 |
| `LM-ANTAM-250G` | Logam Mulia ANTAM 250 gram | 622.230.000 |
| `LM-ANTAM-500G` | Logam Mulia ANTAM 500 gram | 1.244.250.000 |
| `LM-ANTAM-1000G` | Logam Mulia ANTAM 1000 gram | 2.488.000.000 |

### `EMAS-KARAT` — 23 produk, satuan gram

| SKU | Nama | Harga |
|---|---|---|
| `EMAS-24K-999` | Emas 24K 99.9% | 2.200.000 |
| `EMAS-24K-99` | Emas 24K 99% | 2.170.000 |
| `EMAS-23K` | Emas 23K | 2.020.000 |
| `EMAS-22K` | Emas 22K | 1.922.000 |
| `EMAS-21K` | Emas 21K | 1.825.000 |
| `EMAS-20K` | Emas 20K | 1.738.000 |
| `EMAS-19K` | Emas 19K | 1.651.000 |
| `EMAS-18K-CAP755` | Emas 18K Cap 755 | 1.568.000 |
| `EMAS-18K` | Emas 18K | 1.548.000 |
| `EMAS-17K-CAP750` | Emas 17K Cap 750 | 1.523.000 |
| `EMAS-17K` | Emas 17K | 1.470.000 |
| `EMAS-16K-CAP700` | Emas 16K Cap 700 | 1.418.000 |
| `EMAS-16K` | Emas 16K | 1.383.000 |
| `EMAS-15K` | Emas 15K | 1.283.000 |
| `EMAS-14K` | Emas 14K | 1.197.000 |
| `EMAS-13K` | Emas 13K | 1.112.000 |
| `EMAS-12K` | Emas 12K | 1.026.000 |
| `EMAS-11K` | Emas 11K | 941.000 |
| `EMAS-10K` | Emas 10K | 855.000 |
| `EMAS-9K-CAP420` | Emas 9K Cap 420 | 821.000 |
| `EMAS-9K` | Emas 9K | 766.000 |
| `EMAS-8K-CAP375` | Emas 8K Cap 375 | 729.000 |
| `EMAS-8K` | Emas 8K | 681.000 |

## Yang tidak dikirim API ini

Harga pokok (HPP), rata-rata biaya, margin, dan stok **tidak pernah** ada di
balasan — bukan karena belum sempat ditambahkan, tapi karena itu posisi beli
penjual dan bukan urusan pengunjung landing page. Jangan menunggu kolom itu
muncul.
