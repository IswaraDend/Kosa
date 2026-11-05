# API Kosa v1

Dokumen untuk pihak ketiga — landing page, storefront, atau sistem lain yang
perlu membaca katalog produk sebuah project di Kosa.

API ini **hanya baca**. Tidak ada endpoint untuk membuat order, mengubah stok,
atau menyentuh invoice.

---

## Base URL

```
https://<host-kosa>
```

Simpan base URL sebagai variabel konfigurasi, jangan di-hardcode — host
produksinya bisa berubah.

Untuk pengembangan lokal: `http://localhost:8080`.

---

## Kredensial

Setiap project punya satu atau lebih **secret key** berbentuk:

```
kosa_sk_<43 karakter acak>
```

Minta kunci ini ke pemilik project (dibuat dari dashboard Kosa, menu
**Secret Key**). Satu project boleh punya lebih dari satu kunci aktif, jadi
rotasi bisa dilakukan tanpa membuat integrasi Anda mati dulu.

> **Secret key adalah kredensial server.** Jangan pernah menaruhnya di kode
> frontend, bundle JavaScript, atau repositori publik. API ini memang tidak
> mengizinkan pemanggilan langsung dari browser (lihat bagian CORS).

---

## Alur: tukar kunci, lalu pakai token

Secret key **tidak** dipakai langsung di endpoint data. Tukar dulu jadi access
token berumur pendek.

### 1. `POST /api/v1/token`

Kirim secret key di header.

```bash
curl -X POST https://<host-kosa>/api/v1/token \
  -H "X-API-Key: kosa_sk_xxxxxxxxxxxxxxxxxxxx"
```

Alternatif, kalau klien Anda lebih mudah memakai Bearer:

```bash
curl -X POST https://<host-kosa>/api/v1/token \
  -H "Authorization: Bearer kosa_sk_xxxxxxxxxxxxxxxxxxxx"
```

Respons:

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "project": { "id": 3, "name": "Doctor Beans", "code": "DB-01" }
}
```

`expires_in` dalam detik. Simpan token di memori, pakai ulang sampai hampir
kadaluarsa, lalu tukar lagi. Jangan menukar token di setiap request.

### 2. Panggil endpoint data

```bash
curl https://<host-kosa>/api/v1/products \
  -H "Authorization: Bearer <access_token>"
```

---

## Endpoint

Semua endpoint daftar menerima `?page=` dan `?per_page=`, dan membalas amplop
yang sama:

```json
{ "data": [...], "total": 42, "page": 1, "per_page": 25, "total_pages": 2 }
```

`total` adalah jumlah baris yang cocok, bukan jumlah baris di `data`.
Permintaan tanpa kedua parameter itu dijawab utuh tanpa paginasi.

### `GET /api/v1/products`

Daftar produk project pemilik kunci.

| Query | Arti |
|---|---|
| `page`, `per_page` | Paginasi |
| `category` | Hanya produk di satu kategori, pakai **kode** kategori — mis. `?category=MIN`. Tidak membedakan huruf besar/kecil. Kode yang tidak dikenal menjawab daftar kosong, bukan 404 |
| `category_id` | Sama, tapi memakai id angka. `category` lebih disarankan: kode tetap sama walau nama kategorinya diubah |
| `q` | Cari di nama atau SKU (case-insensitive) |

```json
{
  "data": [
    {
      "id": 2,
      "sku": "KOPI-200",
      "name": "Kopi 200gr",
      "unit": "gr",
      "default_price": 20000,
      "category": { "id": 4, "code": "MIN", "name": "Minuman" },
      "updated_at": "2026-11-04T15:22:10.5+07:00"
    }
  ],
  "total": 1,
  "page": 1,
  "per_page": 25,
  "total_pages": 1
}
```

`category` bernilai `null` untuk produk tanpa kategori.

Harga pokok (HPP), rata-rata biaya, dan margin **tidak pernah** dikirim lewat
API ini — itu posisi beli penjual, bukan urusan pembeli.

### `GET /api/v1/products/:id`

Satu produk, bentuknya sama dengan satu elemen `data` di atas.

Produk milik project lain dijawab `404`, bukan `403`.

### `GET /api/v1/categories`

Daftar kategori produk, untuk membangun filter di sisi Anda tanpa harus
menelusuri seluruh produk.

```json
{ "data": [{ "id": 4, "code": "MIN", "name": "Minuman" }], "total": 1, "page": 1, "per_page": 25, "total_pages": 1 }
```

`code` adalah yang sebaiknya Anda simpan di sisi Anda. Ia tidak berubah ketika
pemilik project mengganti nama kategorinya, sementara `name` bisa berubah
kapan saja dan `id` tidak berarti apa-apa saat dibaca orang.

---

## Kode error

| Status | Artinya | Yang perlu dilakukan |
|---|---|---|
| `401` | Secret key atau access token tidak valid, dicabut, atau kadaluarsa | Tukar token baru; kalau tetap gagal, minta kunci baru |
| `403` | Project tidak aktif, atau modul Produk dimatikan untuk project ini | Hubungi pemilik project |
| `404` | Produk tidak ada, atau milik project lain | — |
| `429` | Melewati batas 60 permintaan per menit per kunci | Tunggu sesuai header `Retry-After`, lalu ulangi |

Pesan error berbahasa Indonesia, dalam bentuk `{"error": "..."}`.

Perhatikan bahwa `401` **tidak** membedakan kunci yang salah, dicabut, atau
kadaluarsa. Itu disengaja — membedakannya akan membuat endpoint ini bisa
dipakai menebak kunci mana yang ada.

---

## Batas pemakaian

60 permintaan per menit per secret key. Respons `429` menyertakan
`Retry-After` dalam detik.

Kalau Anda butuh lebih, bicarakan dengan pemilik project — jangan memakai
beberapa kunci untuk menghindari batas ini.

---

## CORS

API ini dirancang **server-to-server**. Header `X-API-Key` sengaja tidak masuk
daftar header yang diizinkan CORS, jadi memanggilnya langsung dari browser
pengunjung tidak akan berhasil.

Panggil dari backend atau build step landing page Anda, simpan hasilnya, lalu
sajikan ke browser dari sisi Anda sendiri.

---

## Rotasi dan pencabutan

- Pemilik project bisa mencabut kunci kapan saja. Pencabutan berlaku
  **seketika** — access token yang masih hidup pun langsung ditolak, tidak
  menunggu satu jamnya habis.
- Satu project boleh punya beberapa kunci aktif sekaligus. Untuk rotasi tanpa
  downtime: minta kunci baru, ganti konfigurasi Anda, baru minta kunci lama
  dicabut.
- Kalau Anda menduga kunci bocor, minta pencabutan lebih dulu, baru urus
  penggantinya.

---

## Contoh lengkap (Node.js)

```js
const BASE = process.env.KOSA_BASE_URL;
const KEY = process.env.KOSA_SECRET_KEY;

let cached = { token: null, expiresAt: 0 };

async function getToken() {
  if (cached.token && Date.now() < cached.expiresAt - 60_000) return cached.token;

  const res = await fetch(`${BASE}/api/v1/token`, {
    method: 'POST',
    headers: { 'X-API-Key': KEY },
  });
  if (!res.ok) throw new Error(`Gagal menukar token: ${res.status}`);

  const body = await res.json();
  cached = { token: body.access_token, expiresAt: Date.now() + body.expires_in * 1000 };
  return cached.token;
}

export async function listProducts({ category, page = 1 } = {}) {
  const token = await getToken();
  const url = new URL(`${BASE}/api/v1/products`);
  url.searchParams.set('page', page);
  url.searchParams.set('per_page', 50);
  // Kode kategori, bukan id: lihat catatan di bagian /api/v1/categories.
  if (category) url.searchParams.set('category', category);

  const res = await fetch(url, { headers: { Authorization: `Bearer ${token}` } });
  if (res.status === 429) throw new Error('Kena rate limit, coba lagi nanti');
  if (!res.ok) throw new Error(`Gagal mengambil produk: ${res.status}`);

  return res.json();
}
```
