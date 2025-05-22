# ginx

`ginx` is a high-performance HTTP reverse proxy written in Go, leveraging low-level Linux syscalls like `epoll` for efficient asynchronous I/O. Inspired by NGINX, it provides a minimal yet powerful async architecture with configurable load balancing.

> ⚠️ This project is a work in progress (WIP). It's being built for learning, performance experimentation, and infrastructure tooling.

---

## ✨ Features

### ✅ Implemented
- **Asynchronous I/O** using `epoll` and raw socket operations
- **HTTP/1.1 Reverse Proxy** with support for common HTTP methods
- **Round-Robin Load Balancing** for distributing traffic across multiple backends
- **Configurable Backends** via YAML configuration
- **Connection Pooling** for efficient resource usage

### ⏳ Planned
- More load balancing algorithms (least connections, IP hash)
- Health checks for backend servers
- Prometheus metrics endpoint
- Dynamic configuration reload

---

## 🚀 Getting Started

### Prerequisites
- Go 1.21 or later
- Linux (uses Linux-specific syscalls)
- Docker (testing environment)

### Installation
```bash
git clone [https://github.com/stanleydv12/ginx.git](https://github.com/stanleydv12/ginx.git)
cd ginx
docker compose up --build -d
```

### Benchmark Methodology

This project includes HTTP performance benchmarks for the `/get` endpoint using [`hey`](https://github.com/rakyll/hey).  
All benchmarks were run locally inside a Docker container with controlled resources.

- **Endpoint:** `http://localhost:8080/get`
- **Total Requests:** 10,000
- **Concurrent Connections:** 50, 500, 1000, 5000

## Benchmark Results

See [`BENCHMARK.md`](./BENCHMARK.md) for detailed results, including summary tables, latency distributions, and histograms for each concurrency level.

---

**Note:**  
Performance may vary depending on hardware, container limits, and OS/network tuning.  
For reproducibility, please use the resource settings above or document your own.


