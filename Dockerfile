FROM golang:1.22-alpine AS builder

WORKDIR /app

# install dependency
RUN apk add --no-cache git ca-certificates

# 🔥 PENTING: biar Go auto download toolchain sesuai go.mod (1.25)
ENV GOTOOLCHAIN=auto
ENV GOPROXY=https://proxy.golang.org,direct

# copy go mod dulu (cache)
COPY go.mod go.sum ./

# download dependency + toolchain
RUN go mod tidy && go mod download

# copy source
COPY . .

# build binary
RUN go build -o app cmd/server/main.go


# =========================
# RUNTIME STAGE
# =========================
FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/app .

EXPOSE 8081

CMD ["./app"]