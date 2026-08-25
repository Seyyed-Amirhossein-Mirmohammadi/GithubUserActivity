# GitHub Events CLI

A robust command-line tool to fetch public events for any GitHub user and store them as a pretty‑printed JSON file.  
It handles retries, rate limits, and uses **smart ETag caching** to avoid redundant downloads.

---

## Features

- **Fetches any GitHub user's public events** (up to 30 most recent).
- **Pretty‑printed JSON output** – saved as `<username>_events.json`.
- **ETag‑based caching** – stores the server’s ETag and validates with `If‑None‑Match`.  
  - Returns `304 Not Modified` when data hasn’t changed – **no wasted bandwidth**.
  - No fixed TTL – the cache is always authoritative.
- **Automatic retries** with exponential backoff for network and server errors (5xx).
- **Rate‑limit awareness** – waits until the reset time when the API quota is exhausted.
- **Handles common errors gracefully** – user not found (404), forbidden (403), etc.
- **Zero external dependencies** – uses only the Go standard library.

---

## Installation

### Build from Source

```bash
git clone https://github.com/Seyyed-Amirhossein-Mirmohammadi/GithubUserActivity
cd github-events-cli
go build -o events ./cmd/events
```

The binary `events` will be created in the current directory.

### Install using `go install`

```bash
go install github.com/yourusername/github-events-cli/cmd/events@latest
```

*(Replace the path with your own if you’ve forked.)*

---

## Usage

```bash
./events <github-username>
```

### Example

```bash
./events Seyyed-Amirhossein-Mirmohammadi
```
 
#### First run (no cache):

```
Fetched fresh events (ETag "W/5e3f4a1b2c3d4e5f6g7h8i9j0k1l2m3n")
Successfully saved fresh events for user 'octocat' to octocat_events.json
```

#### Subsequent run (data unchanged):

```
Using cached events (unchanged, ETag "W/5e3f4a1b2c3d4e5f6g7h8i9j0k1l2m3n")
Successfully saved cached events for user 'octocat' to octocat_events.json
```

#### If the user does not exist:

```
User 'nonexistent' does not exist (404 Not Found).
```

---

## How the Cache Works

The cache is stored in the `cache/` directory inside your working folder.  
For each username, two files are created:

- `<username>.json` – the raw JSON response from the API.
- `<username>.meta` – a JSON file with the following structure:
  ```json
  {
    "timestamp": 1234567890,
    "etag": "W/..."
  }
  ```

The caching flow:

1. On first run for a user, a normal `GET` request is made. The response body and the `ETag` header are stored.
2. On later runs, the client sends `If‑None‑Match: <etag>`.
3. If the data hasn’t changed, GitHub replies with `304 Not Modified` and an empty body. The cached JSON is reused.
4. If the data has changed, GitHub replies with `200 OK`, a new body, and a new ETag. Both are updated in the cache.

This approach **does not rely on a fixed time‑to‑live** – it always serves the latest version available without unnecessary downloads.

---

## Project Structure

```
github-events-cli/
├── go.mod
├── cmd/
│   └── events/
│       └── main.go          # entry point – orchestrates cache, API, and storage
├── internal/
│   ├── cache/
│   │   └── cache.go         # file‑based ETag cache (read/write meta + body)
│   ├── github/
│   │   └── client.go        # HTTP client with retries, rate‑limit handling, and conditional requests
│   └── storage/
│       └── file.go          # JSON pretty‑printing and file writing
└── README.md
```

---

## Dependencies

- **Go 1.21** or later (uses `context`, `slices`, and other standard library features).
- No third‑party dependencies.

---

## Contributing

Contributions are welcome! Feel free to open issues or pull requests.

---

## The roadmap project URL:
https://roadmap.sh/projects/github-user-activity
