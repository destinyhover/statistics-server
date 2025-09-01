# -------- build stage --------
FROM golang:1.24-bookworm AS builder
WORKDIR /src

COPY go.mod ./
RUN go mod download || true

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/app ./...

# -------- runtime stage --------
FROM alpine:3.20
WORKDIR /app

RUN mkdir -p /data
COPY data/data.json /data/data.json     
ENV DATA_PATH=/data/data.json PORT=:1234
VOLUME ["/data"]

COPY --from=builder /out/app /app/app

EXPOSE 1234
ENTRYPOINT ["/app/app"]
