docker run -d --name goapp \
  --restart unless-stopped \
  -p 1234:1234 \
  -e DATA_PATH=/data/data.json \
  -v ~/stat-data:/data \
  ghcr.io/destinyhover/statistics-server:latest
