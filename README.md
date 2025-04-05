# Website Load Handling Checker  

This is a simple Go-based tool to test how a website handles multiple concurrent requests. It allows for sending multiple requests to a target URL, optionally masking IP using randomly generated IP addresses  

By default, `load-checker` can efficiently handle up to **100,000 concurrent requests** using Go’s lightweight goroutines. However, due to system resource constraints, a single Go program might hit practical limits around this number. To **bypass this limitation for extremely high loads**, we provide an optional **helper script (`helper.rs`)** written in Rust, which allows spawning multiple independent instances of `load-checker` in parallel  

**If needed for handling more than 100,000 requests at a time, `helper.rs` can be used. Otherwise, it remains completely optional**  

## Features  

- Sends multiple HTTP GET requests to a specified website  
- Allows setting a custom number of requests (or defaults to a random value between 20-32)  
- Supports IP masking via `Forwarded` headers  
- Uses goroutines for concurrent requests (scales up to ~100K)  
- Reports the number of successful and failed requests  
- Supports launching multiple `load-checker` instances via `helper.rs` (for extreme load testing)  

## Installation  

### **Option 1: Download prebuilt binary**  
You can download the latest release from the [Releases](https://github.com/dhr412/web-load-check/releases) page  

1. **Go to the [Releases](https://github.com/dhr412/web-load-check/releases) page**  
2. **Download the binaries** for your operating system:  
   - **`load-checker`** (`.exe` for Windows or the appropriate file for Linux/macOS)  
   - **`helper`** (`helper.exe` for Windows or the appropriate file for Linux/macOS, if needed)  

3. **Make the files executable:**  
   - **Linux/macOS:**  
     ```sh
     chmod +x load-checker helper
     ```  
   - **Windows:** No action needed. The `.exe` files are already executable  

4. **Run the tool:**  
   - **For a single test:**  
     - **Windows:**  
       ```sh
       load-checker.exe -url https://example.com
       ```  
     - **Linux/macOS:**  
       ```sh
       ./load-checker -url https://example.com
       ```  

   - **For extreme load testing (multiple instances):**  
     - **Windows:**  
       ```sh
       helper.exe 5
       ```  
     - **Linux/macOS:**  
       ```sh
       ./helper 5
       ```  
     - The script will prompt for `load-checker` arguments. Enter:  
       ```sh
       -url https://example.com -requests 500000 -mask=true
       ```  
     - This will start **5 instances** of `load-checker`, each handling 500,000 requests  

### **Option 2: Build from source**  

#### **For Go-based load tester (`load-checker`)**  
Ensure you have [Go](https://go.dev/) installed. Then, clone the repository and build the project:  

```sh
git clone https://github.com/dhr412/web-load-check.git  
cd web-load-check  
go build -o load-checker  
```  

#### **For Rust-based helper script (`helper.rs`)**  
Ensure you have [Rust](https://www.rust-lang.org/tools/install) installed. Then, build the helper script:  

```sh
rustc helper.rs -o helper
```  

## Usage  

### **Running a single load test**  

Run the executable with the required flags:  

```sh
./load-checker -url <website_url> [options]
```  

#### **Flags**  

| Flag         | Description                                       | Default           |
|-------------|---------------------------------------------------|-------------------|
| `-url`      | Target website URL (required)                     | None              |
| `-requests` | Number of requests to send                        | 20-32 (random)    |
| `-mask`     | Enable IP masking (fake `X-Forwarded-For`)        | `true`            |
| `-help`     | Show help message                                 | `false`           |

#### **Examples**  

Send 50 requests to `https://example.com` without IP masking:  

```sh
./load-checker -url https://example.com -requests 50 -mask=false  
```  

Send requests with default settings (random 20-32 requests, with IP masking):  

```sh
./load-checker -url https://example.com  
```  

### **Running multiple instances with `helper.rs` (for extreme load testing)**  

**Use this only if sending more than ~100,000 requests at a time is needed**  

The `helper.rs` script launches multiple independent instances of `load-checker`, allowing to **scale beyond Go’s goroutine limit (~100k per program)** by running multiple separate processes in parallel  

#### **Usage:**  

```sh
./helper <num_instances>
```

- `<num_instances>`: Number of concurrent instances to run. Default is **8** if not specified.  

#### **Example:**  

Run **5 instances** of `load-checker` targeting `https://example.com` with 50 requests per instance:  

```sh
./helper 5
```

The script will then prompt for arguments for `load-checker`, where you can enter:  

```sh
-url https://example.com -requests 500000 -mask=true
```

This will start **5 separate processes**, each executing 500,000 requests in parallel.

### **Use cases of `helper.rs`**  

| Use Case                           | Recommended Approach  |
|------------------------------------|----------------------|
| **≤ 100K requests** (single program)  | `load-checker` only |
| **> 100K requests** (high load testing) | Use `helper.rs` to run multiple instances |

## How It Works  

### **Load Checker (`load-checker`)**
1. Parses command-line arguments  
2. Validates the target URL and request count  
3. Uses goroutines to send concurrent HTTP GET requests (up to ~100K)  
4. If enabled, generates a random IP address for masking  
5. Collects and prints request success/failure statistics  

### **Helper Script (`helper.rs`)**
1. Takes an optional number of instances as input (default: **8**)  
2. Prompts for `load-checker` arguments  
3. Spawns multiple independent instances of `load-checker`  
4. Waits for all instances to complete before exiting  

## License  

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
