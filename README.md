# Website Load Handling Checker

A Go-based tool to test how a website handles multiple concurrent requests. It allows for sending multiple requests to a target URL, optionally masking the host computer and maximizing connection overhead.

By default, `wlc` can efficiently handle massive concurrency using goroutines. 

> **Note:** If your request count exceeds 1,000,000, the tool will automatically cap it at 1,000,000 to prevent local system resource exhaustion.

## Features

- Sends multiple HTTP GET requests to a specified website
- Connection Exhaustion: Forces a fresh TCP/TLS handshake for every request by disabling Keep-Alives and HTTP/2, maximizing server load
- Embedded Assets: User-agent lists are compiled directly into the binary (`//go:embed`), requiring zero external file dependencies at runtime
- Allows setting a custom number of requests (or defaults to a random value if set to `<= 0`)
- Supports host masking via `X-Forwarded-For` headers and user-agent rotation
- Optional randomized request ramp-up (to mimic a more natural traffic pattern)
- Cross-platform prebuilt binaries (Windows, Linux, macOS, and Android/Termux)
- Reports the number of successful and failed requests

---

## Installation

### **Option 1: Download prebuilt binary**

1. **Go to the [Releases](https://github.com/dhr412/web-load-check/releases) page**
2. **Download the binaries for your OS:**
3. **Make them executable (Linux/macOS):**

   ```sh
   chmod +x wlc
   ```

4. **Run the tool with your desired flags:**

```sh
# Basic test
./wlc -url https://example.com

# Heavy load test
./wlc -url https://example.com -requests 50000 -mask=true
```

---

### **Option 2: Build from source**

#### Build `wlc`

```sh
git clone https://github.com/dhr412/web-load-check.git
cd web-load-check
go build -o wlc
```

---

## Usage

### **Basic load test**

```sh
./wlc -url <website_url> [options]
```

#### Flags

| Flag               | Description                                    | Default            |
|--------------      |------------------------------------------------|--------------------|
| `-url`/`-u`        | Target website URL (required)                  | None               |
| `-requests`/`-r`   | Number of requests to send                     | Random 20–32       |
| `-mask`            | Enable IP masking and user-agent rotation      | `true`             |
| `-ranramp`         | Enable randomized ramp-up traffic pattern      | `false`            |
| `-help`            | Show help message                              | N/A                |

---

### **Examples**

Run 50 requests without masking:

```sh
./wlc -url https://example.com -requests 50 -mask false
```

Use default random (20–32) requests with masking:

```sh
./wlc -url https://example.com
```

Use randomized ramp-up pattern:

```sh
./wlc -url https://example.com -requests 5000 -ranramp
```

---

## How It Works

1. Parses CLI arguments and validates inputs using Go's native `flag` package
2. Parses the embedded `user_agents.txt` file into memory at startup
3. Generates randomized or uniform traffic using goroutines
4. Applies IP masking (`X-Forwarded-For`) and randomized headers if enabled
5. Maximizes Server Load: Explicitly disables HTTP Keep-Alives and HTTP/2 multiplexing. This forces the target server to perform a full TCP (and TLS) handshake and teardown for every single request, exhausting connection tables.
6. Supports ramped-up requests using randomized burst batches
7. Tracks and prints request statistics (successful vs unsuccessful)


## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
