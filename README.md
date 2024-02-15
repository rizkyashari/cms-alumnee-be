# Backend Service for LMS Al-Muddatsiriyah

## Pendahuluan

Sistem Backend ini ditulis dengan menggunakan bahasa [Go](https://go.dev) dengan bergantung (_dependent_) dengan [PostgreSQL](https://www.postgresql.org) untuk Database dan [Redis](https://redis.io) sebagai cache.

### Petunjuk Instalasi

Sebelum melakukan instalasi, anda *harus* menyiapkan file `.env`. File `.env` adalah file yang berisi konfigurasi dan kredensial rahasia yang dibutuhkan untuk sistem anda.

Untuk membuat file tersebut, anda dapat memanggil _command_ berikut, pastikan anda berada di "root" directory dari sistem ini:

```bash
touch .env
```

Setelah itu, anda dapat meng-_copy_ isi dari file `.env.example`, dan meletakkannya di file `.env` yang baru anda buat. Penjelasan dari setiap _key_ di `.env` adalah sebagai berikut:

```env
PORT=8081           # Port tempat menjalankan BE (10.xx.xx:8081)
                                                           ^^^^
DB_HOST=pg          # Host untuk Database*
DB_USER=admin       # Username untuk Admin Database*
DB_PASSWORD=admin   # Password untuk Admin Database
DB_NAME=pg          # Nama Database
DB_PORT=5432        # Port dari Host untuk Database*

REDIS_HOST=redis        # Host untuk Redis*
REDIS_PASSWORD=admin    # Password untuk Redis*

JWT_DURATION_MIN=1440               # Durasi JWT (Login)
JWT_SECRET=lT0sVtsavoQkkwW2h9if     # String random untuk JWT

DEFAULT_PASSWORD=password123        # Password default untuk akun baru
```

Nilai dengan tanda `*` tidak perlu diganti jika anda menggunakan Docker. Sedangkan nilai untuk kredensial seperti `DB_USER`, `DB_PASSWORD`, `REDIS_PASSWORD`, `JWT_SECRET`, dan `DEFAULT_PASSWORD` harus diubah agar sistem lebih terjaga.

#### Menggunakan Docker (Recommended)

Sistem ini dapat diinstal dengan menggunakan [Docker](https://www.docker.com). Docker adalah sebuah _container_ yang dapat menjalankan aplikasi secara ter-virtualisasi.

Untuk dapat menjalankan aplikasi menggunakan Docker, pertama-tama pastikan sistem memiliki Docker terinstall dengan menjalankan _command_ berikut di terminal:

```bash
docker -v
```

Jika Docker telah terinstall, maka command tersebut akan mengeluarkan output: `Docker version 24.x.x, build xxx`. Jika tidak, anda dapat mengikuti tutorial [Install Docker Engine](https://docs.docker.com/engine/install/).

Docker telah "membungkus" (_container_) semua kebutuhan dari sistem, seperti DB dan Cache. Sehingga, setelah dipastikan bahwa Docker terinstall, anda dapat melakukan [Database Migration](https://en.wikipedia.org/wiki/Schema_migration) dengan memanggil _command_ berikut di terminal:

```bash
make migrate
```

Setelah proses selesai, anda dapat melakukan [Database Seeding](https://en.wikipedia.org/wiki/Database_seeding), yaitu mengisi data awal pada Database dengan menggunakan _command_ berikut:

```bash
make seed
```

Terakhir, anda dapat menjalankan sistem dengan menggunakan "dev mode" dengan memanggil:

```bash
make dev
```

Atau mode produksi (_production_) dengan memanggil:

```bash
make deploy
```