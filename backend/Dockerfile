# --------------------------------------------------------
# STAGE 1: Builder (Tempat kompilasi codingan Go)
# --------------------------------------------------------
FROM golang:1.26.0-alpine AS builder

# Install git (kadang butuh buat fetch dependency tertentu)
RUN apk add --no-cache git

# Set folder kerja di dalam container
WORKDIR /app

# Copy file dependency dulu (biar dicache sama Docker layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh codingan source code
COPY . .

# Build aplikasi menjadi binary bernama 'server'
# CGO_ENABLED=0 artinya kita build pure static binary (biar jalan di Alpine kosong)
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/property-service/main.go

# --------------------------------------------------------
# STAGE 2: Runner (Image asli yang akan dideploy, size kecil)
# --------------------------------------------------------
FROM alpine:latest

# Install sertifikat SSL (PENTING: biar bisa request HTTPS ke Google/AWS)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary 'server' dari Stage 1 tadi
COPY --from=builder /app/server .
COPY --from=builder /app/.env.example .env 

# Expose port aplikasi
EXPOSE 8080

# Jalankan aplikasi
CMD ["./server"]