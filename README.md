# JbearP Blog

![Lint Status](https://github.com/Jeanpigi/blog/actions/workflows/lint.yml/badge.svg)
![Security Scan](https://github.com/Jeanpigi/blog/actions/workflows/trivy.yml/badge.svg)

A personal blog built with Go — a digital garden for thoughts on technology, programming, and stories. Includes a synchronized live radio broadcast served directly from the server.

## Features

- Blog posts with categories, reading time, and pagination
- Synchronized radio broadcast — all listeners hear the same song at the same position
- Server-side auto-advance: songs change automatically based on their real MP3 duration, no client required
- Visit tracking with geolocation (async, non-blocking)
- Admin dashboard: create, edit, and delete posts; upload MP3s
- PJAX navigation — page transitions without full reload, audio never interrupts
- Rate limiting on login (5 attempts / 15 min per IP)
- Gzip compression, security headers, CSRF protection

## Requirements

- Go 1.23+
- MySQL 8+
- A `./music/` folder for MP3 files (created automatically on first upload)

## Environment Variables

Copy or create a `.env` file in the project root:

```env
SESSION_AUTH_KEY=<base64 of 64 random bytes>
SESSION_ENC_KEY=<base64 of 32 random bytes>
MYSQL_USER=<username>
MYSQL_ROOT_PASSWORD=<password>
MYSQL_HOST=<host>
MYSQL_DATABASE=<database>
MYSQL_PORTDATABASE=3306
ALLOWED_ORIGIN=https://yourdomain.com
```

Generate secure keys:

```bash
openssl rand -base64 64   # SESSION_AUTH_KEY
openssl rand -base64 32   # SESSION_ENC_KEY
```

> The `.env` file is loaded only in `main.go`. Never load `godotenv` in other packages.

## Installation

```bash
git clone git@github.com:Jeanpigi/blog.git
cd blog
go mod tidy
```

## Running

```bash
# Development
go run main.go

# Production (build first)
go build -o main main.go
./main
```

The server starts on port `8080` by default.

## Project Structure

```
.
├── main.go                    # Entry point: env, DB, routes, middleware
├── db/db.go                   # All SQL queries
├── internal/
│   ├── handlers/              # One file per route
│   │   ├── streamHandler.go   # Radio broadcast + server-side auto-advance
│   │   ├── nowPlayingHandler.go
│   │   └── uploadHandler.go
│   ├── middleware/            # RequireAuth, TrackVisitMiddleware
│   ├── models/                # Typed structs (Post, Visit, etc.)
│   ├── music/
│   │   ├── loader.go          # Scans ./music/ folder at startup
│   │   └── duration.go        # MP3 duration parser (Xing/CBR, no dependencies)
│   ├── playlist/playlist.go   # Shuffled playlist with anti-repeat
│   └── utils/                 # Template cache, security helpers
├── session/store.go           # gorilla/sessions (HttpOnly, Secure, SameSiteStrict)
├── templates/                 # HTML templates with shared layout
├── static/                    # CSS, JS, assets
│   └── js/radio.js            # Client sync: polls now-playing, seeks to elapsed position
└── music/                     # MP3 files (gitignored, persists on server)
```

## Radio Architecture

The radio uses a **server-side broadcast model**:

1. On startup, `InitBroadcast()` picks the first song and reads its duration via `music.Duration()` (parses Xing/Info VBR header or estimates from bitrate for CBR files).
2. A `time.AfterFunc` timer fires when the song ends and automatically advances to the next one — completely independent of connected clients.
3. Clients call `GET /api/radio/now-playing` to get `{song, startedAt}`, set `audio.src = /radio/stream`, and seek to `elapsed = (Date.now() - startedAt) / 1000` to synchronize with the broadcast.
4. When a client's audio ends, it polls `now-playing` until `startedAt` changes (server advanced), then syncs to the new song.
5. Admin can skip the current song via `POST /api/radio/advance`.

New MP3 files uploaded via `/radio/upload` are added to the playlist immediately — no restart needed.

## Production Deployment

### systemd service

Create `/etc/systemd/system/blog.service`:

```ini
[Unit]
Description=Blog Project
After=network.target

[Service]
EnvironmentFile=/path/to/blog/.env
ExecStart=/path/to/blog/main
WorkingDirectory=/path/to/blog
User=<user>
Group=<group>
Restart=always
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable blog
sudo systemctl start blog
```

Protect the `.env` file:

```bash
chmod 600 /path/to/blog/.env
```

### nginx reverse proxy

```nginx
server {
    listen 80;
    server_name yourdomain.com;

    # Audio stream: disable buffering to avoid playback cuts
    location /radio/stream {
        proxy_pass http://127.0.0.1:8080;
        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 3600s;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

> `proxy_buffering off` on `/radio/stream` is required — without it nginx buffers the audio before sending it, causing silence or cuts on the client.

## Useful Commands

```bash
# Run tests
go test ./...

# Lint
golangci-lint run

# Tidy dependencies
go mod tidy

# Diagnose DB connection
go run ./cmd/dbcheck/

# Check service logs
journalctl -u blog -n 50 --no-pager
```

## Database

Expected tables: `Users`, `Posts`, `Visits`.

Connection pool: 25 max open/idle connections, 5-minute lifetime.

## CI/CD

- `.github/workflows/lint.yml` — golangci-lint on push/PR to main
- `.github/workflows/trivy.yml` — Trivy security scan (CRITICAL/HIGH, exits 1 on findings)

## Author

**Jean Pierre Arenas**

## License

[MIT License](https://opensource.org/licenses/MIT)
