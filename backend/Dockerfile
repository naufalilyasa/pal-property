# Gunakan image Golang berbasis Debian Trixie
FROM golang:1.26.0-trixie

# Update repository dan install git serta tzdata menggunakan apt
# --no-install-recommends digunakan untuk menjaga ukuran image tetap kecil
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

# Set zona waktu ke Jakarta
ENV TZ=Asia/Jakarta

# Set folder kerja di dalam container
WORKDIR /app

# Install 'air' secara global di dalam container untuk hot-reload
RUN go install github.com/air-verse/air@latest

# Copy file dependency untuk di-cache
# Pastikan file go.sum sudah ada di root project pal-property kamu
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh source code
COPY . .

# Expose port aplikasi sesuai standar Fiber kamu
EXPOSE 8080

# Saat container nyala, otomatis jalankan air!
CMD ["air"]