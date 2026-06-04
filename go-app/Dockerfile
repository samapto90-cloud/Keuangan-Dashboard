FROM golang:1.25-alpine AS builder

WORKDIR /src
RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /keuangan .

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -S sipkeu && adduser -S sipkeu -G sipkeu

WORKDIR /app
COPY --from=builder /keuangan ./keuangan
COPY Anggaran.xlsx ./Anggaran.xlsx

ENV PORT=3000
ENV ANGGARAN_FILE=/app/Anggaran.xlsx
ENV TZ=Asia/Jakarta

USER sipkeu
EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:3000/health || exit 1

CMD ["./keuangan"]
