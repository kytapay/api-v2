# How To Run - KytaPay API v2

Dokumentasi lengkap untuk menjalankan KytaPay API v2 di VPS dengan setup domain, SSL, dan reverse proxy.

**Catatan**: API v2 dan Webhook v2 adalah sistem terpisah dengan repository GitHub terpisah. Mereka bisa dijalankan di VPS yang sama dengan port berbeda (API: 8080, Webhook: 8081).

## 📋 Daftar Isi

1. [Persyaratan](#persyaratan)
2. [Setup VPS](#setup-vps)
3. [Install Dependencies](#install-dependencies)
4. [Setup Database](#setup-database)
5. [Clone & Setup Project](#clone--setup-project)
6. [Setup Environment Variables](#setup-environment-variables)
7. [Build & Run Aplikasi](#build--run-aplikasi)
8. [Setup Systemd Services](#setup-systemd-services)
9. [Setup Domain & DNS](#setup-domain--dns)
10. [Setup Nginx Reverse Proxy](#setup-nginx-reverse-proxy)
11. [Setup SSL dengan Let's Encrypt](#setup-ssl-dengan-lets-encrypt)
12. [Monitoring & Logs](#monitoring--logs)
13. [Troubleshooting](#troubleshooting)

---

## Persyaratan

- **VPS**: Ubuntu 20.04/22.04 LTS (minimal 1GB RAM, 1 CPU core)
- **Domain**: Domain atau subdomain untuk API
  - Contoh: `api.kytapay.com`
- **Database**: MySQL/MariaDB (bisa shared hosting cPanel atau VPS terpisah)
- **Akses**: Root atau sudo access ke VPS
- **Port**: 8080 (untuk API, bisa diubah via environment variable)

---

## Setup VPS

### 1. Update System

```bash
sudo apt update
sudo apt upgrade -y
```

### 2. Buat User Non-Root (Opsional tapi Recommended)

```bash
# Buat user baru
sudo adduser kytapay
sudo usermod -aG sudo kytapay

# Switch ke user baru
su - kytapay
```

### 3. Setup Firewall

```bash
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp    # HTTP untuk Let's Encrypt
sudo ufw allow 443/tcp   # HTTPS
sudo ufw enable
```

---

## Install Dependencies

### 1. Install Docker & Docker Compose

```bash
# Update package index
sudo apt update

# Install prerequisites
sudo apt install -y apt-transport-https ca-certificates curl gnupg lsb-release

# Add Docker's official GPG key
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

# Set up stable repository
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker Engine
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Add user to docker group (untuk menjalankan docker tanpa sudo)
sudo usermod -aG docker $USER
newgrp docker

# Verify installation
docker --version
docker compose version
```

### 2. Install Nginx

```bash
sudo apt install nginx -y
sudo systemctl enable nginx
sudo systemctl start nginx
```

### 3. Install Certbot (untuk SSL)

```bash
sudo apt install certbot python3-certbot-nginx -y
```

### 3. Install Git

```bash
sudo apt install git -y
```

---

## Setup Database

### Jika menggunakan cPanel Shared Hosting:

Ikuti panduan di `DATABASE_SETUP.md` untuk mendapatkan:
- DB_HOST
- DB_PORT
- DB_USER
- DB_PASSWORD
- DB_NAME

### Jika menggunakan MySQL di VPS:

```bash
# Install MySQL Server
sudo apt install mysql-server -y

# Secure installation
sudo mysql_secure_installation

# Login ke MySQL
sudo mysql -u root -p

# Buat database dan user
CREATE DATABASE kytapay_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'kytapay_user'@'localhost' IDENTIFIED BY 'your_secure_password';
GRANT ALL PRIVILEGES ON kytapay_db.* TO 'kytapay_user'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

---

## Clone & Setup Project

### 1. Clone Repository

```bash
# Buat direktori untuk aplikasi
sudo mkdir -p /opt
sudo chown $USER:$USER /opt
cd /opt

# Clone repository API
git clone https://github.com/kytapay/api-v2.git
cd api-v2

# Atau jika sudah ada, pull terbaru
cd /opt/api-v2
git pull origin main

# Pastikan go.sum ada dan lengkap
ls -la go.sum
# Jika go.sum tidak ada atau kosong, jalankan:
go mod tidy
go mod verify
```

### 2. Verify Docker Setup

```bash
# Test Docker
docker run hello-world

# Check Docker Compose
docker compose version
```

---

## Setup Environment Variables

### 1. Setup Environment Variables

```bash
cd /opt/api-v2
cp env.example .env
nano .env
```

Isi dengan konfigurasi yang sesuai (lihat `env.example` untuk referensi).

**PENTING**: Pastikan file `.env` ada di direktori `/opt/api-v2` sebelum menjalankan Docker Compose. Docker akan mount file ini ke dalam container.

### 2. Setup RSA Keys (jika diperlukan)

```bash
# PakaiLink RSA Private Key (untuk API)
cd /opt/api-v2
# Upload file pkcs8_rsa_private_key.pem ke direktori ini
# Pastikan permission-nya aman
chmod 600 pkcs8_rsa_private_key.pem
```

---

## Build & Run dengan Docker

### 1. Setup Environment Variables

Pastikan file `.env` sudah dibuat di kedua folder (lihat bagian Setup Environment Variables sebelumnya).

### 2. Build Docker Image

```bash
cd /opt/api-v2

# Pastikan .env file ada
if [ ! -f .env ]; then
    echo "ERROR: .env file tidak ditemukan!"
    echo "Copy env.example ke .env dan isi dengan konfigurasi yang sesuai"
    exit 1
fi

# Pastikan go.sum ada dan lengkap sebelum build
if [ ! -f go.sum ] || [ ! -s go.sum ]; then
    echo "go.sum tidak ditemukan atau kosong, menjalankan go mod tidy..."
    go mod tidy
    go mod verify
fi

# Build image untuk API
docker compose build

# Jika masih error, coba build tanpa cache
# docker compose build --no-cache
```

### 3. Run dengan Docker Compose

```bash
cd /opt/api-v2

# Start service
docker compose up -d

# Check status
docker compose ps

# View logs
docker compose logs -f

# View logs untuk service
docker compose logs -f api-v2
```

### 4. Test Service

```bash
# Test API health
curl http://localhost:8080/health
```

### 5. Useful Docker Commands

```bash
# Stop service
docker compose stop

# Start service
docker compose start

# Restart service
docker compose restart

# Stop and remove container
docker compose down

# Stop, remove container, and remove volumes
docker compose down -v

# Rebuild and restart
docker compose up -d --build

# View logs
docker compose logs -f

# Execute command in container
docker compose exec api-v2 sh

# Check container resource usage
docker stats
```

---

## Setup Docker untuk Auto-Start (Systemd)

### 1. Enable Docker Service

```bash
# Enable Docker to start on boot
sudo systemctl enable docker
sudo systemctl start docker
```

### 2. Create Systemd Service untuk Docker Compose (Opsional)

Jika ingin menggunakan systemd untuk manage Docker Compose:

```bash
sudo nano /etc/systemd/system/kytapay-docker.service
```

Isi dengan:

```ini
[Unit]
Description=KytaPay Docker Compose Services
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/opt/api-v2
ExecStart=/usr/bin/docker compose up -d
ExecStop=/usr/bin/docker compose down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
```

Enable service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable kytapay-docker.service
sudo systemctl start kytapay-docker.service
```

### 3. Alternative: Docker Restart Policy

Docker Compose sudah menggunakan `restart: unless-stopped`, jadi container akan otomatis restart jika crash atau server reboot. Tidak perlu systemd service tambahan.

### 4. Useful Commands

```bash
cd /opt/api-v2

# Check Docker service status
sudo systemctl status docker

# View Docker Compose logs
docker compose logs -f

# View logs untuk service tertentu
docker compose logs -f api-v2
docker compose logs -f webhook-v2

# Restart services
docker compose restart

# Restart specific service
docker compose restart api-v2
docker compose restart webhook-v2

# Stop services
docker compose stop

# Start services
docker compose start

# Remove and recreate containers
docker compose up -d --force-recreate
```

---

## Setup Domain & DNS

### 1. Point Domain ke VPS

Di panel DNS domain Anda, tambahkan A record:

```
Type: A
Name: api (atau @ untuk root domain)
Value: [IP_VPS_ANDA]
TTL: 3600
```

Contoh:
- `api.kytapay.com` → `123.456.789.0`

### 2. Verify DNS

```bash
# Cek apakah DNS sudah propagate
dig api.kytapay.com

# atau
nslookup api.kytapay.com
```

Tunggu beberapa menit/jam sampai DNS propagate (biasanya 5-30 menit).

---

## Setup Nginx Reverse Proxy

### 1. Create Nginx Config

```bash
sudo nano /etc/nginx/sites-available/kytapay-api
```

Copy isi dari `deploy/nginx/kytapay-api.conf` atau isi dengan:

```nginx
server {
    listen 80;
    server_name api-v2.kytapay.com;

    # Logging
    access_log /var/log/nginx/kytapay-api-access.log;
    error_log /var/log/nginx/kytapay-api-error.log;

    # Client body size limit
    client_max_body_size 10M;

    # Proxy to API service
    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Health check endpoint
    location /health {
        proxy_pass http://localhost:8080/health;
        access_log off;
    }
}
```

### 2. Enable Site

```bash
# Enable site
sudo ln -s /etc/nginx/sites-available/kytapay-api /etc/nginx/sites-enabled/

# Test Nginx configuration
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx
```

---

## Setup SSL dengan Let's Encrypt

### 1. Install SSL

```bash
sudo certbot --nginx -d api-v2.kytapay.com
```

Ikuti instruksi:
- Enter email address
- Agree to terms
- Choose whether to redirect HTTP to HTTPS (pilih 2 untuk redirect)

### 2. Auto-Renewal

Certbot sudah setup auto-renewal, tapi bisa test manual:

```bash
# Test renewal
sudo certbot renew --dry-run

# Check renewal status
sudo certbot certificates
```

### 3. Update Nginx Config (Setelah SSL)

Nginx config akan otomatis di-update oleh Certbot. File akan ada di:
- `/etc/nginx/sites-available/kytapay-api` (dengan SSL)

---

## Monitoring & Logs

### 1. View Application Logs

```bash
cd /opt/api-v2

# View all logs
docker compose logs -f

# API logs
docker compose logs -f api-v2
docker compose logs api-v2 --since 1h

# View last 100 lines
docker compose logs --tail=100 api-v2
```

### 2. View Nginx Logs

```bash
# API access logs
sudo tail -f /var/log/nginx/kytapay-api-access.log
sudo tail -f /var/log/nginx/kytapay-api-error.log
```

### 3. Check Service Status

```bash
cd /opt/api-v2

# Check Docker containers status
docker compose ps

# Check specific container
docker ps | grep kytapay-api

# Check container health
docker inspect kytapay-api-v2 | grep -A 10 Health

# Check if port is listening
sudo netstat -tlnp | grep 8080

# Check Nginx status
sudo systemctl status nginx

# Check Docker service
sudo systemctl status docker
```

### 4. Health Check

```bash
# Test API health
curl https://api.kytapay.com/health
# atau
curl http://localhost:8080/health
```

---

## Troubleshooting

### 1. Container Tidak Start

```bash
cd /opt/api-v2

# Check container logs
docker compose logs api-v2

# Check container status
docker compose ps

# Check if container is running
docker ps -a | grep kytapay-api

# Check container exit code
docker inspect kytapay-api-v2 | grep ExitCode

# Check .env file
cat /opt/api-v2/.env

# Try to start container manually
docker compose up api-v2
```

### 1a. Docker Build Error (go.sum missing)

Jika error saat build seperti "missing go.sum entry":

```bash
cd /opt/api-v2

# Pull latest code
git pull origin main

# Pastikan go.sum ada
ls -la go.sum

# Jika go.sum tidak ada atau kosong, generate ulang
go mod tidy
go mod verify

# Commit dan push go.sum (jika perlu)
git add go.sum go.mod
git commit -m "Update go.sum"
git push origin main

# Build ulang
docker compose build --no-cache
```

### 1b. .env File Not Found Error

Jika muncul pesan "No .env file found, using environment variables":

```bash
cd /opt/api-v2

# Pastikan .env file ada
ls -la .env

# Jika tidak ada, copy dari env.example
if [ ! -f .env ]; then
    cp env.example .env
    nano .env  # Edit dengan konfigurasi yang sesuai
fi

# Pastikan permission benar
chmod 600 .env

# Restart container untuk reload .env
docker compose down
docker compose up -d

# Check logs
docker compose logs api-v2
```

**Catatan**: Docker Compose menggunakan `env_file` untuk load environment variables, tapi `godotenv` juga mencari file `.env` di working directory. File `.env` sudah di-mount ke `/root/.env` di container, jadi aplikasi akan menemukannya.

### 2. Database Connection Error

```bash
# Test database connection
mysql -h [DB_HOST] -u [DB_USER] -p[DB_PASSWORD] [DB_NAME] -e "SELECT 1;"

# Check firewall
sudo ufw status

# Check database credentials di .env
```

### 3. Port Already in Use

```bash
# Check what's using port 8080
sudo lsof -i :8080
sudo lsof -i :8081

# Kill process if needed
sudo kill -9 [PID]
```

### 4. Nginx 502 Bad Gateway

```bash
# Check if container is running
docker ps | grep kytapay-api

# Check Nginx error logs
sudo tail -f /var/log/nginx/kytapay-api-error.log

# Test proxy manually
curl http://localhost:8080/health
```

### 5. SSL Certificate Issues

```bash
# Check certificate status
sudo certbot certificates

# Renew manually
sudo certbot renew

# Check Nginx SSL config
sudo nginx -t
```

### 6. Permission Issues

```bash
# Fix ownership untuk .env file
sudo chown $USER:$USER /opt/api-v2/.env

# Fix permissions untuk .env file
sudo chmod 600 /opt/api-v2/.env

# Fix permissions untuk RSA key
sudo chmod 600 /opt/api-v2/pkcs8_rsa_private_key.pem

# Check Docker permissions
sudo usermod -aG docker $USER
newgrp docker
```

### 7. Docker Issues

```bash
# Check Docker daemon
sudo systemctl status docker

# Restart Docker
sudo systemctl restart docker

# Check Docker Compose version
docker compose version

# Check disk space
df -h

# Clean up Docker (remove unused images, containers, volumes)
docker system prune -a

# Check container resource usage
docker stats
```

---

## Update Aplikasi

### 1. Pull Latest Code

```bash
cd /opt/api-v2
git pull origin main

# Pastikan go.sum tetap ada setelah pull
if [ ! -f go.sum ] || [ ! -s go.sum ]; then
    echo "go.sum hilang setelah pull, menjalankan go mod tidy..."
    go mod tidy
    go mod verify
fi
```

### 2. Rebuild Docker Image

```bash
cd /opt/api-v2

# Rebuild image
docker compose build

# Rebuild without cache (jika ada masalah)
docker compose build --no-cache
```

### 3. Restart Service

```bash
cd /opt/api-v2

# Restart dengan rebuild
docker compose up -d --build

# Atau restart saja (jika tidak ada perubahan code)
docker compose restart

# Restart specific service
docker compose restart api-v2
```

### 4. Zero-Downtime Update (Recommended)

```bash
# Pull latest code
cd /opt/api-v2
git pull origin main

# Rebuild image
docker compose build

# Rolling update (stop old, start new)
docker compose up -d --no-deps --build api-v2
```

---

## Security Best Practices

1. **Jangan commit file sensitif ke Git**
   - `.env` files
   - RSA keys (`.pem` files)
   - Pastikan sudah di `.gitignore`

2. **Set Permission yang Tepat**
   ```bash
   chmod 600 .env
   chmod 600 pkcs8_rsa_private_key.pem
   chmod 644 pakailink_rsa_public_key.pem
   ```

3. **Gunakan Firewall**
   ```bash
   sudo ufw enable
   sudo ufw status
   ```

4. **Update System Regularly**
   ```bash
   sudo apt update && sudo apt upgrade -y
   ```

5. **Monitor Logs**
   - Setup log rotation
   - Monitor untuk suspicious activity

---

## File yang TIDAK BOLEH di-push ke Git

Pastikan file-file berikut ada di `.gitignore`:

- `.env` (environment variables)
- `pkcs8_rsa_private_key.pem` (PakaiLink private key)
- `pakailink_rsa_public_key.pem` (PakaiLink public key)
- `*.log` (log files)
- Binary files (`api-v2`, `webhook-v2`)

---

## Support

Jika ada masalah, cek:
1. Logs aplikasi: `sudo journalctl -u kytapay-api.service -f`
2. Logs Nginx: `sudo tail -f /var/log/nginx/kytapay-api-error.log`
3. Service status: `sudo systemctl status kytapay-api.service`

---

**Selamat! Aplikasi KytaPay API v2 dan Webhook v2 sudah berjalan di VPS Anda! 🎉**

