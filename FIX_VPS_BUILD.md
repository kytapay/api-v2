# Fix Docker Build Error di VPS

Jika Anda mendapat error saat build Docker di VPS seperti:

```
missing go.sum entry for module providing package...
```

## Solusi Cepat

### 1. Pull Latest Code

```bash
cd /opt/api-v2
git pull origin main
```

### 2. Pastikan go.sum Ada dan Lengkap

```bash
# Cek apakah go.sum ada
ls -la go.sum

# Jika go.sum tidak ada atau kosong, generate ulang
go mod tidy
go mod verify
```

### 3. Build Ulang Docker

```bash
# Build dengan no-cache untuk memastikan fresh build
docker compose build --no-cache
```

## Jika Masih Error

### Option 1: Install Go di VPS (Recommended)

```bash
# Install Go (jika belum ada)
cd /tmp
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify
go version

# Generate go.sum
cd /opt/api-v2
go mod tidy
go mod verify
```

### Option 2: Update Dockerfile untuk Auto-Generate

Jika Go tidak tersedia di VPS, update Dockerfile untuk auto-generate go.sum:

```dockerfile
# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies and generate go.sum if missing
RUN go mod download
RUN if [ ! -f go.sum ] || [ ! -s go.sum ]; then go mod tidy; fi
RUN go mod verify
```

Tapi cara terbaik adalah memastikan `go.sum` sudah di-commit dan di-push ke GitHub.

## Verifikasi

Setelah fix, pastikan:

```bash
# go.sum ada dan tidak kosong
ls -lh go.sum

# go.sum valid
go mod verify

# Build berhasil
docker compose build
```

