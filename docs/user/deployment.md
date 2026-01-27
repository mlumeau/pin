# Deployment

This guide focuses on a simple binary deployment.

## Build
```bash
go build -o pin
```

## Run
```bash
PIN_ENV=production \
PIN_SECRET_KEY=change-me \
PIN_BASE_URL=https://example.com \
PIN_DB_PATH=/var/lib/pin/identity.db \
PIN_UPLOADS_DIR=/var/lib/pin/uploads \
PIN_COOKIE_SECURE=true \
./pin
```

## systemd example
```ini
[Unit]
Description=PIN
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/pin
ExecStart=/opt/pin/pin
Restart=on-failure
Environment=PIN_ENV=production
Environment=PIN_SECRET_KEY=change-me
Environment=PIN_BASE_URL=https://example.com
Environment=PIN_DB_PATH=/var/lib/pin/identity.db
Environment=PIN_UPLOADS_DIR=/var/lib/pin/uploads
Environment=PIN_COOKIE_SECURE=true

[Install]
WantedBy=multi-user.target
```

## Reverse proxy (Caddy)
```
example.com {
  reverse_proxy 127.0.0.1:5000
}
```

## Reverse proxy (nginx)
```
server {
  listen 80;
  server_name example.com;
  location / {
    proxy_pass http://127.0.0.1:5000;
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
  }
}
```

## Notes
- Keep `WorkingDirectory` set to the repo root so `templates/` and `static/` resolve.
- Ensure `cwebp` is installed on the host for WebP encoding.
- When using a reverse proxy with HTTPS, set `PIN_BASE_URL` to the public URL and `PIN_COOKIE_SECURE=true`.
- If `PIN_BASE_URL` is unset, OAuth callbacks may be incorrect behind a proxy.
- Ensure `PIN_UPLOADS_DIR` and the database path are writable by the service user.
