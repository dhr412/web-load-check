# Website Load Handling Checker  

This is a simple Go-based tool to test how a website handles multiple concurrent requests. It allows you to send multiple requests to a target URL, optionally masking your IP using randomly generated IP addresses.  

## Features  

- Sends multiple HTTP GET requests to a specified website  
- Allows setting a custom number of requests (or defaults to a random value between 20-32)  
- Supports IP masking via `Forwarded` headers  
- Uses goroutines for concurrent requests  
- Reports the number of successful and failed requests  

## Installation  

### **Option 1: Download Prebuilt Binary**  
You can download the latest release from the [Releases](https://github.com/dhr412/website-load-checker/releases) page  

1. **Go to the [Releases](https://github.com/dhr412/website-load-checker/releases) page**  
2. **Download the binary** for your operating system (`.exe` for Windows, or the appropriate file for Linux/macOS)  
3. **Make the file executable:**  
   - **Linux/macOS:**  
     ```sh
     chmod +x load-checker
     ```  
   - **Windows:** No action needed. The `.exe` file is already executable.  

4. **Run the tool:**  
   - **Windows:**  
     ```sh
     load-checker.exe -url https://example.com
     ```  
     - **Linux/macOS:**  
     ```sh
     ./load-checker -url https://example.com
     ```  

### **Option 2: Build from Source**  
Ensure you have [Go](https://go.dev/) installed. Then, clone the repository:  

```sh
git clone https://github.com/dhr412/website-load-checker.git  
cd website-load-checker  
go build -o load-checker  
```  

## Usage  

Run the executable with the required flags:  

```sh
./load-checker -url <website_url> [options]
```  

### Flags  

| Flag          | Description                                       | Default |
|--------------|---------------------------------------------------|---------|
| `-url`       | Target website URL (required)                     | None    |
| `-requests`  | Number of requests to send                        | 20-32 (random) |
| `-mask`      | Enable IP masking (fake `X-Forwarded-For`)        | `true`  |
| `-help`      | Show help message                                 | `false` |

### Examples  

Send 50 requests to `https://example.com` without IP masking:  

```sh
./load-checker -url https://example.com -requests 50 -mask=false  
```  

Send requests with default settings (random 20-32 requests, with IP masking):  

```sh
./load-checker -url https://example.com  
```  

## How It Works  

1. Parses command-line arguments  
2. Validates the target URL and request count  
3. Uses goroutines to send concurrent HTTP GET requests  
4. If enabled, generates a random IP address for masking  
5. Collects and prints request success/failure statistics  

## License  

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.  
