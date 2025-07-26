# Website Load Handling Checker

A Go-based tool to test how a website handles multiple concurrent requests. It allows for sending multiple requests to a target URL, optionally masking the host computer.

By default, `load-checker` can efficiently handle up to **100,000 concurrent requests** using goroutines. However, due to system resource constraints, a single Go program might hit practical limits around this number.

> **Note:** If your request count exceeds 100,000, the tool will automatically cap it at 100,000.

## Features

- Sends multiple HTTP GET requests to a specified website
- Allows setting a custom number of requests (or defaults to a random value between 20-32)
- Supports host masking via headers
- Optional randomized request ramp-up (to mimic a more natural traffic pattern)
- Uses goroutines for concurrency (scales up to ~100K)
- Reports the number of successful and failed requests

---

## Installation

### **Option 1: Download prebuilt binary**

1. **Go to the [Releases](https://github.com/dhr412/web-load-check/releases) page**
2. **Download the binaries** for your OS:
3. **Make them executable (Linux/macOS):**

   ```sh
   chmod +x load-checker
   ```

4. **Run the script:**

#### For a single test

```sh
./load-checker -url https://example.com
```

Enter arguments when prompted:

```sh
-url https://example.com -requests 500000 -mask true
```

---

### **Option 2: Build from source**

#### Build `load-checker`

```sh
git clone https://github.com/dhr412/web-load-check.git
cd web-load-check
go build -o load-checker
```

---

## Usage

### **Basic load test**

```sh
./load-checker -url <website_url> [options]
```

#### Flags

| Flag         | Description                                    | Default            |
|--------------|------------------------------------------------|--------------------|
| `-url`       | Target website URL (required)                  | None               |
| `-requests`  | Number of requests to send                     | Random 20–32       |
| `-mask`      | Enable IP masking and user-agent rotation      | `true`             |
| `-ranramp`   | Enable randomized ramp-up traffic pattern      | `false`            |
| `-help`      | Show help message                              | `false`            |

---

### **Examples**

Run 50 requests without masking:

```sh
./load-checker -url https://example.com -requests 50 -mask false
```

Use default random (20–32) requests with masking:

```sh
./load-checker -url https://example.com
```

Use randomized ramp-up pattern:

```sh
./load-checker -url https://example.com -requests 5000 -ranramp
```

---

## How It Works

### `load-checker`

1. Parses CLI arguments and validates inputs
2. Generates randomized or uniform traffic
3. Uses goroutines to send HTTP GET requests in parallel
4. Applies IP masking and randomized headers if enabled
5. Supports ramped-up requests using randomized burst batches
6. Tracks and prints request statistics

## License  

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
