# streamgate

A small Go service that adds **accounts, stream keys and publish authorization** to a self-hosted [MediaMTX](https://github.com/bluenviron/mediamtx) live-streaming setup.

MediaMTX moves the video. streamgate decides **who is allowed to broadcast** and keeps track of **what is live**.

> **Status: early.** It works end to end on a local machine (OBS → MediaMTX → VLC), but it has not been security-audited and has no web UI yet. Please read [Known gaps](#known-gaps) before using it for anything real.

> **Intended use:** streaming content you own or have permission to broadcast. You are responsible for what is streamed through any server you run.

## What it does

- Registers users and logs them in (bcrypt-hashed passwords, JWT access tokens).
- Creates streams. Each stream gets a secret **stream key** (used to broadcast) and a separate public **playback ID** (used to watch).
- Answers MediaMTX's "may this publish?" callback, so only someone holding a valid key can broadcast.
- Lists the streams that are live, and keeps live/offline status in sync with MediaMTX automatically.

## How it works

```
                  publish (RTMP :1935)
  OBS / encoder ──────────────────────────► ┌────────────┐
                                            │  MediaMTX  │ ──► viewers (HLS :8888)
                                            └─────┬──────┘
                                                  │   ▲
        "may this key publish?"                   │   │  "which paths are live?"
        POST /internal/mediamtx/auth              ▼   │  GET :9997/v3/paths/list (every 3s)
                                            ┌────────────┐
                                            │ streamgate │ ──► PostgreSQL
                                            │    API     │     (users, streams)
                                            └────────────┘
```

What happens when someone presses **Start Streaming** in OBS:

1. OBS connects to MediaMTX and sends the stream's playback ID and stream key.
2. MediaMTX asks the streamgate API whether that key may publish to that path.
3. The API looks the key up in PostgreSQL and answers `200` (allowed) or `401` (refused).
4. Video flows. MediaMTX slices it into HLS for viewers.
5. Every few seconds the API asks MediaMTX which paths are live and updates each stream's status, so a stream goes offline by itself when the broadcast stops, even if the encoder crashes.

**Why the key travels as a "password".** In MediaMTX the path you publish to is also the path viewers watch. If the stream key were the path, the viewer URL would reveal it. So the **public playback ID is the path**, and the **secret stream key is sent as the password** in the connection URL:

```
OBS server:      rtmp://localhost
OBS stream key:  <playback_id>?user=stream&pass=<stream_key>
Viewers watch:   http://localhost:8888/<playback_id>/index.m3u8
```

## Quick start (about 10 minutes)

### Prerequisites

| Tool | Used for |
|---|---|
| Go (a recent release, see `go.mod`) | building and running the API |
| PostgreSQL | users and streams |
| [MediaMTX](https://github.com/bluenviron/mediamtx/releases) | ingest and HLS delivery |
| [OBS Studio](https://obsproject.com/) | the broadcaster |
| [VLC](https://www.videolan.org/) | a simple player for testing |

Tested with: Go `<version>`, PostgreSQL `<version>`, MediaMTX `<version>`, on Windows with Git Bash. Other platforms should work with minor changes.

### 1. Get the code

```bash
git clone https://github.com/<your-username>/streamgate.git
cd streamgate
go mod download
```

### 2. Create the database

Create an empty database, then run the schema in `migrations/001_init.sql` against it.

```bash
psql -U postgres -c "CREATE DATABASE streamingserver;"
psql -U postgres -d streamingserver -f migrations/001_init.sql
```

No `psql`? Create the database in pgAdmin, open a Query Tool on it, paste the contents of `migrations/001_init.sql` and run it.

### 3. Configure the API

```bash
cp .env.example .env
```

Edit `.env` and set your database details and a JWT secret. Generate a secret with:

```bash
openssl rand -base64 32
```

The `.env` file is read from the folder you run the API from, and it is in `.gitignore`. Never commit it.

### 4. Configure MediaMTX

Open `mediamtx.yml` (next to the MediaMTX binary) and set these values:

```yaml
authMethod: http
authHTTPAddress: http://127.0.0.1:8095/internal/mediamtx/auth
authHTTPExclude:
  - action: api
  - action: metrics
  - action: pprof

api: true
apiAddress: 127.0.0.1:9997
```

Use spaces, not tabs, and make sure there is only one `authMethod` line. MediaMTX only reads the file at startup, so restart it after editing.

Optional: if VLC will not play the stream, set `hlsVariant: mpegts` for maximum player compatibility.

### 5. Run both programs

In one terminal, start the API (leave it running):

```bash
go run ./cmd/api
```

Wait for `Listening and serving HTTP on :8095`. In a second terminal, start MediaMTX:

```bash
./mediamtx
```

## Demo: your first stream

### 6. Create a user and a stream

Run this in a third terminal. On Windows, use Git Bash.

```bash
API=http://127.0.0.1:8095

# Register
curl -s -X POST $API/register -H "Content-Type: application/json" \
  -d '{"username":"demo","email":"demo@example.com","password":"change-me-123"}'

# Log in and keep the token
TOKEN=$(curl -s -X POST $API/login -H "Content-Type: application/json" \
  -d '{"email":"demo@example.com","password":"change-me-123"}' \
  | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

# Create a stream
RESP=$(curl -s -X POST $API/streams -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" -d '{"title":"My first stream"}')
KEY=$(echo "$RESP" | grep -o '"stream_key":"[^"]*"' | cut -d'"' -f4)
PB=$(echo "$RESP" | grep -o '"playback_id":"[^"]*"' | cut -d'"' -f4)

echo "OBS server:      rtmp://localhost"
echo "OBS stream key:  $PB?user=stream&pass=$KEY"
echo "VLC address:     http://localhost:8888/$PB/index.m3u8"
```

The last three lines print exactly what to paste into OBS and VLC. **The stream key is shown only once**, when the stream is created.

### 7. Broadcast from OBS

1. Open **Settings → Stream**.
2. **Service:** Custom
3. **Server:** `rtmp://localhost` (nothing after `localhost`)
4. **Stream key:** the whole `OBS stream key` line printed above, with no spaces
5. Add a source under **Sources**: *Display Capture* for your screen, or *Video Capture Device* for a camera.
6. Under **Settings → Output**, a video bitrate around 2500 Kbps and a keyframe interval of 2 seconds work well for 720p.
7. Click **Start Streaming**. A green square at the bottom of OBS means it is connected.

### 8. Watch it in VLC

Open **Media → Open Network Stream** (Ctrl+N), paste the `VLC address` printed above, and click Play. Wait 10 to 20 seconds for the first picture. A delay of a few seconds behind real life is normal for HLS.

### 9. Check that status follows reality

```bash
curl -s $API/streams          # your stream, with "status":"live" and no stream key
```

Now click **Stop Streaming** in OBS and run the same command again after a few seconds. The list should be empty (`[]`).

## API reference

| Method | Path | Auth | Purpose |
|---|---|---|---|
| `POST` | `/register` | none | Create an account |
| `POST` | `/login` | none | Exchange email and password for a JWT |
| `GET` | `/streams` | none | List live streams (stream keys are never included) |
| `POST` | `/streams` | `Authorization: Bearer <token>` | Create a stream. Returns the stream key once |
| `POST` | `/internal/mediamtx/auth` | localhost only | Callback used by MediaMTX. Not for clients |

Request bodies:

```json
// POST /register   (username: at least 3 characters, password: at least 8)
{ "username": "demo", "email": "demo@example.com", "password": "change-me-123" }

// POST /login
{ "email": "demo@example.com", "password": "change-me-123" }
// → 200 { "token": "eyJ..." }

// POST /streams    (requires Bearer token)
{ "title": "My first stream" }
// → 201 { "id": 1, "title": "My first stream", "stream_key": "live_...", "playback_id": "pb_..." }
```

Typical responses: `201` created, `400` invalid input, `401` missing or wrong credentials, `409` email or username already taken.

## Configuration

All settings are environment variables, loaded from `.env` if present. See `.env.example`.

| Variable | Description | Example |
|---|---|---|
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | PostgreSQL user | `postgres` |
| `DB_PASSWORD` | PostgreSQL password | `change-me` |
| `DB_NAME` | Database name | `streamingserver` |
| `DB_SSLMODE` | `disable` for local development | `disable` |
| `JWT_SECRET` | Secret used to sign login tokens. Keep it private | output of `openssl rand -base64 32` |

Ports used:

| Port | Program | Exposure |
|---|---|---|
| `8095` | streamgate API | keep private unless you add HTTPS and a reverse proxy |
| `1935` | MediaMTX RTMP (broadcasters) | open to the broadcasters you want to allow |
| `8888` | MediaMTX HLS (viewers) | open to viewers, ideally behind HTTPS |
| `9997` | MediaMTX control API | localhost only |
| `5432` | PostgreSQL | never expose to the internet |

The API port is currently set in `cmd/api/main.go`. It was chosen to avoid clashing with other local web servers.

## Project layout

```
cmd/api/          entry point: connects the database and wires the layers together
internal/
  auth/           JWT, auth middleware, stream key and playback ID generators
  database/       PostgreSQL connection (sqlx + pgx)
  mediamtx/       small client for MediaMTX's control API
  stream/         streams: model, repository, service, HTTP handlers, status sync
  user/           users: register and login (bcrypt)
migrations/       SQL schema
```

Each feature package follows the same pattern: `model.go` (data), `repository.go` (the only code that touches the database), `service.go` (rules), `handler.go` (HTTP).

## Security notes

- Passwords are hashed with bcrypt and never returned by the API.
- The **stream key** (secret, lets you broadcast) is separate from the **playback ID** (public, lets you watch). The key is never included in `GET /streams`.
- The MediaMTX callback only accepts requests from the same machine. Keep the API and MediaMTX on the same host, or adapt the check before splitting them.
- Do not expose ports `8095`, `9997` or `5432` to the internet.
- Failed publish attempts are logged for debugging. Do not run with verbose logging in production, since logs may contain mistyped credentials.

## Troubleshooting

| Symptom | Likely cause and fix |
|---|---|
| OBS refused; MediaMTX logs `failed to authenticate ... 401` | The **Stream key** field must hold the whole `playback_id?user=stream&pass=stream_key` string, not only the `live_...` key. The **Server** field must be exactly `rtmp://localhost` |
| API log shows `path="live_..."` | Same cause: only the key was entered, so MediaMTX used it as the path |
| MediaMTX logs `500` from the auth call | Read the `publish refused: ... err=...` line in the API log for the real reason |
| `curl: (7) Failed to connect` to the API | The API is not running. On Windows, if `go run` stops by itself, build and run it: `go build -o api.exe ./cmd/api` then `./api.exe` |
| A `404` HTML page from `127.0.0.1:8095` or `8081` | Another web server (for example Apache from XAMPP) owns that port. Change the API port in `cmd/api/main.go` and in `authHTTPAddress` |
| Stream stays `live` after stopping OBS, or `status sync:` errors in the API log | MediaMTX's control API is off. Set `api: true` and `apiAddress: 127.0.0.1:9997`, then restart MediaMTX |
| VLC shows nothing | Wait 10 to 20 seconds. Check the playback ID in the address. Try `hlsVariant: mpegts` in `mediamtx.yml` |

## Known gaps

This is an early project. Known limitations:

- **No web UI.** Everything is done with `curl`, OBS and VLC. A minimal viewer page is a planned addition.
- **Stream keys are stored in plain text** in the database and cannot be regenerated or revoked yet. Hashing them, and adding rotate and delete endpoints, are planned.
- **No endpoint to list your own streams**, so the key shown at creation is the only chance to copy it.
- **Viewing is open.** Anyone who knows a stream's playback ID can watch it. There are no private or paid streams yet.
- **No rate limiting** on `/register` and `/login`, no email verification, no password reset.
- **Tokens:** one signing secret, 24-hour expiry, no refresh tokens and no revocation.
- **Status sync polls every 3 seconds.** If MediaMTX itself is unreachable, statuses are left unchanged until it is back.
- **Single quality, single node.** The broadcaster's stream is passed through as is: no transcoding to multiple bitrates, no recording, no chat, no scaling across servers.
- **The API port is hard-coded** in `cmd/api/main.go`.
- **No HTTPS configuration included.** Browsers require HTTPS pages to load HTTPS video, so put a reverse proxy such as Caddy in front before real use.
- **Automated tests** are not in place yet.

## Contributing

Issues and pull requests are welcome. For anything larger than a small fix, please open an issue first so we can agree on the approach.

## License

MIT. See [LICENSE](LICENSE).
