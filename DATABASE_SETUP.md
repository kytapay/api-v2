# Database Setup untuk cPanel Shared Hosting

## Cara Mendapatkan Informasi Database dari cPanel

1. **Login ke cPanel**
   - Masuk ke cPanel hosting Anda
   - Cari section "Databases" atau "MySQL Databases"

2. **Buat Database (jika belum ada)**
   - Klik "MySQL Databases"
   - Buat database baru (contoh: `kytw4343_pay`)
   - Buat user baru untuk database tersebut
   - Berikan semua privileges ke user tersebut

3. **Informasi yang Diperlukan**

   **DB_HOST:**
   - Biasanya: `localhost`
   - Atau: `mysql.yourdomain.com`
   - Atau: IP address yang diberikan cPanel
   - Cek di cPanel > MySQL Databases > "Current Host" atau "Remote MySQL"

   **DB_PORT:**
   - Default: `3306`
   - Biasanya tidak perlu diubah untuk shared hosting

   **DB_USER:**
   - Format: `cpanel_username_dbuser`
   - Contoh: Jika cPanel username adalah `kytw4343` dan database user adalah `payuser`
   - Maka DB_USER: `kytw4343_payuser`

   **DB_PASSWORD:**
   - Password yang Anda set saat membuat MySQL user di cPanel

   **DB_NAME:**
   - Format: `cpanel_username_dbname`
   - Contoh: Jika cPanel username adalah `kytw4343` dan database name adalah `pay`
   - Maka DB_NAME: `kytw4343_pay`

## Contoh Konfigurasi

Berdasarkan screenshot yang diberikan (database: `kytw4343_pay`):

```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=kytw4343_payuser
DB_PASSWORD=your-secure-password-here
DB_NAME=kytw4343_pay
```

## Testing Koneksi

Setelah mengisi `.env`, test koneksi dengan menjalankan aplikasi. Jika ada error, cek:

1. **Host salah**: Coba ganti `localhost` dengan IP atau hostname yang diberikan cPanel
2. **User/Password salah**: Pastikan username dan password sesuai dengan yang di cPanel
3. **Database tidak ada**: Pastikan database sudah dibuat di cPanel
4. **Remote access**: Jika aplikasi di server berbeda, pastikan IP server diizinkan di "Remote MySQL" di cPanel

## Troubleshooting

### Error: "Access denied for user"
- Pastikan username dan password benar
- Pastikan user memiliki privileges untuk database tersebut

### Error: "Unknown database"
- Pastikan nama database benar (termasuk prefix cPanel username)
- Pastikan database sudah dibuat di cPanel

### Error: "Can't connect to MySQL server"
- Cek apakah DB_HOST benar (coba `localhost` atau hostname dari cPanel)
- Cek apakah port 3306 tidak diblokir firewall
- Untuk remote connection, pastikan IP diizinkan di "Remote MySQL"

