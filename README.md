# Website Load Handling Checker

This is a simple Go-based tool to test how a website handles multiple concurrent requests. It allows for sending multiple requests to a target URL, optionally masking the host computer.

By default, `load-checker` can efficiently handle up to **100,000 concurrent requests** using Go’s lightweight goroutines. However, due to system resource constraints, a single Go program might hit practical limits around this number. To **bypass this limitation for extremely high loads**, we provide an optional **helper script (`helper.rs`)** written in Rust, which allows spawning multiple independent instances of `load-checker` in parallel.

> **Note:** If your request count exceeds 100,000, the tool will automatically cap it at 100,000 unless using the helper.

## Features

- Sends multiple HTTP GET requests to a specified website  
- Allows setting a custom number of requests (or defaults to a random value between 20-32)  
- Supports host masking via headers  
- Supports keep-alive connections  
- Optional randomized request ramp-up (to mimic a more natural traffic pattern)  
- Uses goroutines for concurrency (scales up to ~100K)  
- Reports the number of successful and failed requests  
- Supports launching multiple `load-checker` instances via `helper` (for extreme load testing)

---

## Installation

### **Option 1: Download prebuilt binary**

1. **Go to the [Releases](https://github.com/dhr412/web-load-check/releases) page**
2. **Download the binaries** for your OS:
   - `load-checker`
   - `helper`
3. **Make them executable (Linux/macOS):**
   ```sh
   chmod +x load-checker helper
   ```

4. **Run the script:**

#### For a single test:

```sh
./load-checker -url https://example.com
```

#### For high load with multiple instances (requires Rust helper):

```sh
./helper ./load-checker 5
```

Enter arguments when prompted:

```sh
-url https://example.com -requests 500000 -mask=true
```

---

### **Option 2: Build from source**

#### Build `load-checker` (Go)

```sh
git clone https://github.com/dhr412/web-load-check.git
cd web-load-check
go build -o load-checker
```

#### Build `helper.rs` (Rust)

```sh
rustc helper.rs -o helper
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
| `-keepalive` | Use persistent connections                     | `true`             |
| `-ranramp`   | Enable randomized ramp-up traffic pattern      | `false`            |
| `-help`      | Show help message                              | `false`            |

---

### **Examples**

Run 50 requests without masking:

```sh
./load-checker -url https://example.com -requests 50 -mask=false
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

### **Running multiple instances with `helper`**

Use `helper` to scale beyond 100K requests.

```sh
./helper ./load-checker <num_instances>
```

- First argument: Path to `load-checker` binary  
- Second argument: Number of instances to launch (default: 8 if omitted)

#### Example:

```sh
./helper ./load-checker 4
```

**When prompted:**

```sh
-url https://example.com -requests 500000 -mask=true
```

This will launch 4 separate processes, each handling 500,000 requests.

---

## Use Cases

| Load Type           | Recommendation                |
|---------------------|-------------------------------|
| ≤ 100K requests      | Use `load-checker` directly   |
| > 100K requests      | Use `helper` to parallelize   |

---

## How It Works

### `load-checker` (Go)

1. Parses CLI arguments and validates inputs  
2. Generates randomized or uniform traffic  
3. Uses goroutines to send HTTP GET requests  
4. Applies IP masking and randomized headers if enabled  
5. Supports ramped-up requests using randomized burst batches  
6. Tracks and prints request statistics

### `helper` (Rust)

1. Requires `load-checker` path and optional instance count  
2. Prompts user for arguments to be passed  
3. Spawns independent `load-checker` processes in threads  
4. Waits for all processes to complete

## License  

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
