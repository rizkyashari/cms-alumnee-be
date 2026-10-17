# Skenario uji: Agenda alumni (R10) dan hak akses terkait (R01)

Dokumen ini dipakai untuk memeriksa fitur agenda di staging terhadap requirement:

> **R10 — Agenda.** Berisi agenda Alumni Pakci Delayota (member verified bisa
> meninggalkan komentar untuk berinteraksi). Member verified bisa submit draft.

Pemetaan ke spreadsheet `implementation-plan-and-daily-checklist.xlsx`:

| Tugas | Isi | Bagian di bawah |
| --- | --- | --- |
| 77 | Daftar/detail agenda published | B, G |
| 78 | Submit draft oleh verified member dan review admin | D |
| 79 | Komentar agenda khusus verified member | E |
| 80 | Uji tanggal, workflow, permission, layout responsif | C, D, F, H |
| R01 | Hak akses dibatasi di backend, per role | C |

Fitur tambahan yang ikut diuji: pendaftaran peserta (bagian F), data asli
menggantikan mock di beranda/ringkasan/dashboard (bagian G), audit (bagian I).

## A. Persiapan

**Akun** (lihat [staging-walkthrough.md](staging-walkthrough.md) untuk cara
mengaktifkannya; semua memakai satu password `STAGING_SCENARIO_PASSWORD`).

| Peran di skenario | Username | Syarat |
| --- | --- | --- |
| Pengunjung anonim | (tanpa login, jendela incognito) | - |
| Admin | `admin.uji` | - |
| Pengurus | `pengurus.satu` | - |
| Member verified A | `alumni.satu` | Jadi Verified lewat langkah 2 walkthrough (bootstrap) |
| Member verified B | `alumni.dua` | Jadi Verified lewat langkah 3-4 walkthrough (5 attestasi + Admin) |
| Member non-verified | `alumni.tujuh` | Tetap Non-verified (jangan diajukan verifikasi) |

Siapkan **dua browser/profil** (atau satu biasa + satu incognito) supaya bisa
login dua akun bersamaan, terutama untuk bagian E dan F.

**Catatan pengujian yang perlu diketahui sebelum mulai**

1. Waktu ditampilkan dalam **WIB**; input tanggal di form memakai zona waktu
   browser. Pastikan jam laptop/browser benar.
2. Agenda **tidak bisa diajukan jika waktu mulainya sudah lewat**. Untuk menguji
   "Sudah lewat" dan "tandai hadir", buat agenda yang mulai **sekitar 5-10 menit
   dari sekarang**, lalu tunggu waktunya tiba (lihat skenario F.8 dan G.6).
3. Pesan error dari API berbahasa Inggris (contoh: `forbidden`,
   `comment must be 1-1000 characters`). Itu perilaku yang diharapkan saat ini.
4. `pengurus.satu` dan `admin.uji` adalah akun fixture **tanpa profil alumni**.
   Mereka tidak bisa mendaftar hadir (lihat F.4); itu bukan bug.
5. Batas komentar: **10 komentar per menit per akun**.

**Data awal.** Pastikan `/agenda` kosong (belum ada agenda terbit) sebelum
mulai, atau catat agenda yang sudah ada agar tidak tertukar.

Tulis hasil tiap skenario di kolom **Hasil** (`OK` / `GAGAL` / `N/A`) dan catat
bukti (screenshot atau pesan error) untuk yang gagal.

---

## B. Pengunjung anonim: baca agenda dan komentar (tugas 77)

Akun: anonim (incognito). Prasyarat: minimal satu agenda terbit (jalankan D
sampai D.14 dulu, lalu kembali ke sini).

| No | Langkah | Hasil yang diharapkan | Hasil |
| --- | --- | --- | --- |
| B.1 | Buka `/agenda` saat belum ada agenda terbit | Tampil tanpa error, pesan "Belum ada agenda mendatang untuk filter ini." | |
| B.2 | Setelah ada agenda terbit, muat ulang `/agenda` | Kartu agenda tampil: tanggal (hari + bulan), kategori, judul, jadwal + lokasi, "N terdaftar · N komentar" | |
| B.3 | Klik filter kategori (mis. Reuni) | Hanya agenda kategori itu; "Semua" mengembalikan semua | |
| B.4 | Ketik sebagian judul di kotak "Cari agenda" | Daftar menyaring setelah jeda singkat (sekitar 0,3 detik) | |
| B.5 | Ketik `%` atau `_` di kotak cari | Tidak error dan tidak mengembalikan semua agenda (karakter dianggap teks biasa) | |
| B.6 | Pindah ke tab **Sudah lewat** lalu kembali ke **Mendatang** | Tab Sudah lewat hanya berisi agenda yang sudah selesai; Mendatang berisi yang belum | |
| B.7 | Klik **Detail** pada satu agenda | Dialog tampil: judul, kategori, jadwal WIB, lokasi, deskripsi, "Diajukan oleh ..." | |
| B.8 | Lihat bagian komentar di dialog tanpa login | **Komentar terbaca** (daftar komentar tampil), tidak ada kotak tulis, ada pesan "Masuk sebagai alumni terverifikasi untuk menulis komentar." dengan tautan Masuk | |
| B.9 | Lihat bagian pendaftaran tanpa login | Terlihat "N alumni terdaftar" dan tautan "Masuk untuk mendaftar hadir."; tidak ada tombol daftar | |
| B.10 | Di tab Developer Tools → Network, perhatikan permintaan komentar | `GET /api/v1/events/{id}/comments` status 200; tidak ada tombol Ubah/Hapus di komentar mana pun | |
| B.11 | Buka `/api/v1/events/ID-ASAL` (ID acak, bukan agenda) di browser | 404 `event not found` | |
| B.12 | Coba akses agenda berstatus draf/menunggu review lewat ID-nya (ambil dari Agenda saya) | 404, **tidak** tampil publik | |

---

## C. Hak akses per peran (R01, dibatasi di backend)

Uji juga langsung ke API agar tidak hanya mengandalkan menu yang disembunyikan.
Contoh cara (ganti `<URL>` dan cookie sesi dari browser; atau pakai DevTools →
Network → "Copy as cURL" dari permintaan yang berhasil lalu ubah akunnya):

```bash
curl -i -X POST https://<URL>/api/v1/events \
  -H 'Content-Type: application/json' -H 'Origin: https://<URL>' \
  -b 'COOKIE_SESI' -d '{"title":"Uji","category":"reuni","description":"x","starts_at":"2030-01-01T10:00:00Z"}'
```

| No | Aksi | Anonim | Non-verified | Verified | Pengurus | Admin |
| --- | --- | --- | --- | --- | --- | --- |
| C.1 | Lihat daftar/detail agenda terbit | Ya | Ya | Ya | Ya | Ya |
| C.2 | Baca komentar | Ya | Ya | Ya | Ya | Ya |
| C.3 | Tulis komentar | 401 | **403** | Ya | Ya | Ya |
| C.4 | Simpan draf agenda (`POST /events`) | 401 | **403** | Ya | Ya | Ya |
| C.5 | Daftar hadir | 401 | Ya* | Ya* | Ya* | Ya* |
| C.6 | Antrean review (`GET /admin/events`) | 401 | 403 | 403 | 403 | Ya |
| C.7 | Terbitkan/tolak/arsipkan | 401 | 403 | 403 | 403 | Ya |
| C.8 | Daftar peserta + tandai hadir | 401 | 403 | Hanya jika penulis | Hanya jika penulis | Ya |

\* Hanya jika akun punya profil alumni; kalau tidak, API menjawab 409 (lihat F.4).

Cek tambahan:

| No | Langkah | Hasil yang diharapkan | Hasil |
| --- | --- | --- | --- |
| C.9 | Login `alumni.tujuh`; buka menu sidebar | **Tidak ada** menu "Agenda saya" | |
| C.10 | Sebagai `alumni.tujuh`, buka langsung `/panel/alumni/agenda` | Pesan "Akun ini tidak memiliki izin untuk membuka halaman ini." | |
| C.11 | Sebagai Verified, buka `/panel/admin/agenda` | Ditolak (tidak punya izin) | |
| C.12 | Sebagai Verified, panggil `GET /api/v1/admin/events` lewat cURL | 403 `forbidden` | |
| C.13 | Kirim `POST /api/v1/events` dengan header `Origin: https://situs-lain.example` | 403 `request origin is not allowed` | |
| C.14 | Login `pengurus.satu`; lihat sidebar | Ada menu **Agenda** (`/panel/pengurus/agenda`) | |
| C.15 | Login `admin.uji`; lihat sidebar | Ada **Kelola agenda** | |
| C.16 | Logout, lalu tekan Back di browser dan coba aksi di panel | Diarahkan ke login / aksi ditolak 401 | |

---

## D. Alur draf dan review (tugas 78)

Akun: `alumni.satu` (penulis, Verified) dan `admin.uji`. Gunakan tanggal **masa
depan** untuk semua agenda.

| No | Akun | Langkah | Hasil yang diharapkan | Hasil |
| --- | --- | --- | --- | --- |
| D.1 | alumni.satu | Buka **Agenda saya** | Halaman tampil, pesan "Kamu belum mengajukan agenda." | |
| D.2 | alumni.satu | Klik **Draf baru**, kosongkan semua lalu cek tombol **Simpan draf** | Tombol nonaktif sampai judul dan deskripsi terisi | |
| D.3 | alumni.satu | Isi judul 2 huruf, deskripsi, mulai, simpan | Ditolak: pesan `title must be 3-200 characters` | |
| D.4 | alumni.satu | Isi **Selesai** lebih awal dari **Mulai**, simpan | Ditolak (pesan "Waktu selesai harus setelah waktu mulai." atau error API) | |
| D.5 | alumni.satu | Isi benar: judul "Temu kangen lintas angkatan", kategori Reuni, lokasi, mulai +3 hari 10.00, selesai 12.00, deskripsi berisi `<b>tebal</b>`; simpan | Tersimpan, muncul kartu dengan chip **Draf** dan kategori | |
| D.6 | alumni.satu | Cek teks deskripsi yang berisi `<b>` | Tampil sebagai **teks biasa** `<b>tebal</b>`, bukan huruf tebal | |
| D.7 | (anonim) | Cek `/agenda` | Draf **tidak** tampil | |
| D.8 | alumni.satu | Klik **Ubah** pada draf, ganti lokasi, simpan | Perubahan tersimpan | |
| D.9 | alumni.satu | Buat draf kedua, lalu **Hapus** (konfirmasi) | Draf hilang dari daftar | |
| D.10 | alumni.satu | Buat draf dengan waktu mulai **di masa lalu**, lalu klik **Ajukan** | Ditolak: `event must start in the future` | |
| D.11 | alumni.satu | Klik **Ajukan** pada draf yang benar | Notifikasi "Agenda dikirim untuk ditinjau Admin."; chip berubah **Menunggu review**; tombol Ubah/Ajukan/Hapus hilang | |
| D.12 | admin.uji | Buka **Kelola agenda**, tab **Menunggu review** | Agenda tampil dengan nama pengaju | |
| D.13 | admin.uji | Klik **Tolak**, kosongkan alasan lalu OK | Ditolak: "Alasan penolakan wajib diisi." | |
| D.14 | admin.uji | Klik **Tolak**, isi "Lengkapi lokasi" | Notifikasi sukses; agenda pindah ke tab **Ditolak** | |
| D.15 | alumni.satu | Muat ulang **Agenda saya** | Chip **Ditolak** dan kotak "Catatan reviewer: Lengkapi lokasi"; tombol Ubah dan Hapus muncul (Ajukan tidak) | |
| D.16 | alumni.satu | **Ubah** agenda yang ditolak lalu simpan | Status kembali **Draf**, catatan reviewer hilang | |
| D.17 | alumni.satu | **Ajukan** lagi | Status **Menunggu review** | |
| D.18 | admin.uji | Klik **Terbitkan** (konfirmasi) | Notifikasi "Agenda diterbitkan."; pindah ke tab **Terbit** | |
| D.19 | (anonim) | Muat ulang `/agenda` | Agenda **tampil publik** | |
| D.20 | alumni.satu | Cek **Agenda saya** | Chip **Terbit**; tidak ada tombol Ubah/Hapus; ada **Peserta (0)** | |
| D.21 | alumni.dua | Login, coba `PUT /api/v1/events/{id}` agenda milik alumni.satu lewat cURL | 404 `event not found` (tidak bisa mengubah agenda orang lain) | |
| D.22 | alumni.satu | Coba `DELETE /api/v1/events/{id}` agenda yang sudah terbit | 409 `event is not in a state that allows this action` | |
| D.23 | admin.uji | Klik **Arsipkan** pada agenda terbit (konfirmasi) | Pindah ke tab **Diarsipkan** | |
| D.24 | (anonim) | Muat ulang `/agenda` | Agenda **hilang** dari publik; buka detailnya lewat ID memberi 404 | |
| D.25 | admin.uji | Buat agenda sendiri lewat bagian "Agenda yang saya ajukan" di halaman Kelola agenda, ajukan, lalu terbitkan | Berhasil (Admin boleh mengajukan dan menerbitkan agenda sendiri) | |
| D.26 | pengurus.satu | Di **Agenda** (pengurus), buat dan ajukan draf | Berhasil; muncul di antrean Admin | |

---

## E. Komentar agenda (tugas 79)

Prasyarat: satu agenda **terbit** (ulangi D sampai D.18, **jangan arsipkan**).
Login bersamaan: `alumni.satu` (browser 1), `alumni.dua` (browser 2),
`admin.uji`, `pengurus.satu`, `alumni.tujuh`.

**Menulis dan label**

| No | Akun | Langkah | Hasil yang diharapkan | Hasil |
| --- | --- | --- | --- | --- |
| E.1 | alumni.satu | Di `/agenda` → Detail, tulis "Siap hadir!" lalu **Kirim komentar** | Komentar muncul; di samping nama ada **centang biru + "Member verified"** | |
| E.2 | pengurus.satu | Tulis komentar | Muncul dengan label **"Pengurus"** (centang hijau) | |
| E.3 | admin.uji | Tulis komentar | Muncul dengan label **"Admin"** (centang ungu) | |
| E.4 | alumni.tujuh | Buka Detail agenda | Komentar **terbaca**, tetapi tidak ada kotak tulis; pesan "Komentar hanya dapat ditulis oleh alumni terverifikasi." | |
| E.5 | alumni.tujuh | Coba `POST /api/v1/events/{id}/comments` lewat cURL | 403 `forbidden` | |
| E.6 | (anonim) | Coba `POST` komentar lewat cURL tanpa cookie | 401 | |
| E.7 | alumni.dua | Cek label nama sendiri | "Member verified" | |

**Validasi dan keamanan**

| No | Akun | Langkah | Hasil yang diharapkan | Hasil |
| --- | --- | --- | --- | --- |
| E.8 | alumni.satu | Kirim komentar berisi spasi saja | Tombol nonaktif / ditolak | |
| E.9 | alumni.satu | Tempel teks 1.001 karakter | Kolom membatasi 1.000 karakter; lewat cURL → 422 `comment must be 1-1000 characters` | |
| E.10 | alumni.satu | Kirim `<script>alert(1)</script> & <b>x</b>` | Tampil sebagai **teks apa adanya**; tidak ada popup | |
| E.11 | alumni.satu | Kirim komentar dengan baris baru (Enter) | Baris baru terjaga saat ditampilkan | |
| E.12 | alumni.satu | Kirim 11 komentar dalam waktu kurang dari satu menit | Komentar ke-11 ditolak: `too many comments, try again shortly` (HTTP 429); setelah 1 menit bisa lagi | |

**Ubah dan hapus**

| No | Akun | Langkah | Hasil yang diharapkan | Hasil |
| --- | --- | --- | --- | --- |
| E.13 | alumni.satu | Di komentar sendiri: **Ubah**, ganti teks, **Simpan** | Teks berubah; muncul keterangan **"· diubah"** | |
| E.14 | alumni.dua | Lihat komentar alumni.satu | **Tidak ada** tombol Ubah/Hapus | |
| E.15 | alumni.dua | Coba `PUT .../comments/{id}` milik alumni.satu lewat cURL | 403 `only the author or a moderator may change this comment` | |
| E.16 | admin.uji | Lihat komentar orang lain | Ada tombol **Ubah** dan **Hapus** | |
| E.17 | admin.uji | **Ubah** komentar milik pengurus.satu menjadi "Dimoderasi" | Berhasil, ada "· diubah" | |
| E.18 | pengurus.satu | Hapus komentar sendiri (konfirmasi) | Komentar hilang; jumlah komentar berkurang | |
| E.19 | admin.uji | Hapus komentar milik alumni.satu | Berhasil | |
| E.20 | admin.uji | Buka **Kelola agenda** → tab Terbit → **Komentar (n)** | Dialog moderasi tampil dengan semua komentar | |

**Kondisi khusus**

| No | Langkah | Hasil yang diharapkan | Hasil |
| --- | --- | --- | --- |
| E.21 | Arsipkan agenda lalu coba tulis komentar baru (cURL sebagai Verified) | 409 `comments are only open on published events` | |
| E.22 | Pada agenda berstatus draf, coba `GET .../comments` sebagai anonim | 404 | |
| E.23 | Muat ulang `/agenda` | Angka "N komentar" di kartu sesuai jumlah komentar | |

---

## F. Pendaftaran peserta

Akun: `alumni.tujuh` (Non-verified, boleh mendaftar), `alumni.satu` (penulis),
`pengurus.satu` (tanpa profil), `admin.uji`.

| No | Akun | Langkah | Hasil yang diharapkan | Hasil |
| --- | --- | --- | --- | --- |
| F.1 | alumni.tujuh | Buka Detail agenda terbit, klik **Daftar hadir** | Chip **"Kamu terdaftar"**, tombol berubah **Batalkan pendaftaran**, angka terdaftar +1 | |
| F.2 | alumni.tujuh | Muat ulang halaman dan buka Detail lagi | Status tetap terdaftar | |
| F.3 | (anonim) | Lihat kartu agenda | "N terdaftar" naik; nama peserta **tidak** terlihat | |
| F.4 | pengurus.satu | Klik **Daftar hadir** | Ditolak: `complete your alumni profile before registering` (409); ini sesuai karena akun fixture tanpa profil | |
| F.5 | alumni.tujuh | Klik **Batalkan pendaftaran** | Status kembali belum terdaftar; angka -1 | |
| F.6 | alumni.tujuh | Daftar lagi setelah batal | Berhasil (pendaftaran dibuka kembali, tidak duplikat) | |
| F.7 | alumni.satu | Di **Agenda saya** → **Peserta (n)** | Daftar berisi nama + angkatan peserta; tombol **Tandai hadir** nonaktif dengan catatan "Kehadiran dapat dicatat setelah agenda dimulai." | |

**Agenda yang sudah mulai (perlu agenda dengan waktu dekat)**

Buat agenda baru sebagai alumni.satu yang mulai **+6 menit** dan selesai
**+30 menit**, ajukan, terbitkan oleh Admin, daftarkan alumni.tujuh. Tunggu
sampai jam mulai lewat.

| No | Akun | Langkah | Hasil yang diharapkan | Hasil |
| --- | --- | --- | --- | --- |
| F.8 | alumni.satu | Buka **Peserta**, klik **Tandai hadir** untuk alumni.tujuh | Status jadi chip **Hadir** | |
| F.9 | alumni.tujuh | Buka Detail agenda itu | Chip **"Kamu hadir"**; **tidak ada** tombol batalkan | |
| F.10 | alumni.tujuh | Coba `DELETE .../registration` lewat cURL | 409 (tidak bisa batal setelah hadir tercatat) | |
| F.11 | alumni.dua | Coba `POST .../attendees/{profileId}/attend` lewat cURL | 403 (bukan penulis/Admin) | |
| F.12 | admin.uji | Buka **Peserta** pada agenda yang sama | Admin juga bisa melihat dan menandai hadir | |
| F.13 | alumni.dua | Setelah agenda **selesai**, buka Detail | Chip "Agenda sudah selesai"; tidak ada tombol daftar; lewat cURL `POST .../registration` → 409 `event has already finished` | |

---

## G. Data asli menggantikan mock

| No | Akun | Langkah | Hasil yang diharapkan | Hasil |
| --- | --- | --- | --- | --- |
| G.1 | anonim | Buka beranda `/`, bagian **Agenda** | Maksimal 3 agenda mendatang **asli** (bukan "Temu kangen lintas angkatan 22 JUN", "Alumni Career Circle #03", "Seminar berbagi untuk masa depan") | |
| G.2 | anonim | Saat belum ada agenda mendatang | Pesan "Belum ada agenda mendatang. Pantau terus halaman ini." | |
| G.3 | anonim | Klik salah satu agenda di beranda | Menuju `/agenda` | |
| G.4 | alumni.tujuh | Buka `/ringkasan` sebelum mendaftar apa pun | Kartu menampilkan **00 agenda terdaftar**; "Agenda berikutnya": "Kamu belum mendaftar ke agenda mendatang." | |
| G.5 | alumni.tujuh | Daftar ke agenda mendatang, muat ulang `/ringkasan` | Angka naik; "Agenda berikutnya" menampilkan agenda itu (judul, jadwal, lokasi) | |
| G.6 | alumni.tujuh | Setelah agenda selesai, cek `/ringkasan` | Agenda tersebut tidak lagi dipilih sebagai "berikutnya" | |
| G.7 | admin.uji | Buka `/panel/admin` | Kartu **"Agenda mendatang"** menampilkan angka asli dan "N menunggu review"; kartu tautan "N agenda mendatang" | |
| G.8 | admin.uji | Ajukan satu agenda dari akun lain, muat ulang dashboard | "menunggu review" bertambah | |
| G.9 | semua | Cari teks "Temu kangen lintas angkatan" yang bukan buatan Anda di seluruh halaman agenda | Tidak ditemukan (mock dihapus) | |

Catatan: angka lain di dashboard/ringkasan (total alumni, berita, usaha,
"teman terhubung", "kontribusi aktif") **belum** dari data asli dan di luar
cakupan agenda.

---

## H. Layout responsif dan aksesibilitas

Uji di tiga lebar: **375 px** (HP), **768 px** (tablet), **1280 px** (desktop).
Gunakan DevTools → Toggle device toolbar.

| No | Halaman | Hasil yang diharapkan | Hasil |
| --- | --- | --- | --- |
| H.1 | `/agenda` | Filter (cari, kategori, tab) tidak meluber; kartu tidak terpotong; tombol Detail terjangkau; tidak ada scroll horizontal | |
| H.2 | Dialog Detail agenda | Dialog muat di layar, isi bisa di-scroll, komentar panjang tidak melebar keluar (teks panjang tanpa spasi pun terpotong rapi) | |
| H.3 | Form draf agenda | Dua kolom tanggal menjadi satu kolom di HP; tombol Simpan terlihat | |
| H.4 | Agenda saya / Kelola agenda | Kartu dan tombol aksi menumpuk rapi di HP | |
| H.5 | Beranda, bagian Agenda | Baris agenda terbaca dan bisa diklik | |
| H.6 | Navigasi keyboard | Tab menjangkau tombol Detail, kotak komentar, tombol Kirim; Esc menutup dialog | |
| H.7 | Teks panjang | Judul agenda 200 karakter dan komentar 1.000 karakter tidak merusak tata letak | |

---

## I. Audit dan data (pemeriksaan di server)

Setelah menjalankan D sampai F, periksa jejak audit di database:

```bash
cd ~/apps/alumnihub/cms-alumnee-be
sudo docker compose exec -T mysql sh -c 'mysql -uroot -p"$MYSQL_ROOT_PASSWORD" alumnee -e "SELECT action, resource_type, created_at FROM audit_events WHERE action LIKE \"event.%\" ORDER BY created_at DESC LIMIT 40;" 2>/dev/null'
```

| No | Pemeriksaan | Hasil yang diharapkan | Hasil |
| --- | --- | --- | --- |
| I.1 | Setiap aksi draf/review tercatat | Ada baris `event.create`, `event.update`, `event.submit`, `event.reject`, `event.publish`, `event.archive`, `event.delete` | |
| I.2 | Komentar tercatat | Ada `event.comment.create`, `event.comment.update`, `event.comment.delete` | |
| I.3 | Pendaftaran tercatat | Ada `event.register`, `event.unregister`, `event.attend` | |
| I.4 | Aksi yang **ditolak** tidak menulis audit | Mis. komentar 403 atau reject tanpa catatan tidak menambah baris | |
| I.5 | Moderasi terlihat | Baris `event.comment.update/delete` oleh Admin pada komentar orang lain punya metadata `moderated: true` | |

Pemeriksaan tabel (opsional):
`SELECT status, COUNT(*) FROM alumni_events GROUP BY status;` harus konsisten
dengan yang terlihat di tab Kelola agenda.

---

## J. Ringkasan hasil

| Bagian | Jumlah skenario | OK | GAGAL | N/A | Catatan |
| --- | --- | --- | --- | --- | --- |
| B Anonim | 12 | | | | |
| C Hak akses | 16 | | | | |
| D Draf dan review | 26 | | | | |
| E Komentar | 23 | | | | |
| F Pendaftaran | 13 | | | | |
| G Data asli | 9 | | | | |
| H Responsif | 7 | | | | |
| I Audit | 5 | | | | |

**Kriteria lulus R10:** B.2 sampai B.10, C.3, C.4, C.6, C.7, D.5 sampai D.19,
dan E.1 sampai E.6 semuanya OK. Ini inti requirement (agenda tampil, member
verified bisa submit draf dan berkomentar, non-verified dan anonim tidak bisa).

**Format laporan bug**

```
ID skenario:
Akun/peran:
Langkah (ringkas):
Hasil diharapkan:
Hasil sebenarnya:
Bukti (screenshot / pesan error / respons API):
Browser + lebar layar:
```

## Reset data uji

Agenda yang dibuat saat uji tetap tersimpan. Untuk membersihkan agenda uji
(komentar dan pendaftaran ikut terhapus karena relasi cascade), jalankan **hanya
di staging**:

```bash
sudo docker compose exec -T mysql sh -c 'mysql -uroot -p"$MYSQL_ROOT_PASSWORD" alumnee -e "DELETE FROM alumni_events;" 2>/dev/null'
```

Audit tetap disimpan.
