# ginx

`ginx` is a lightweight, high-performance HTTP reverse proxy written in Go using low-level Linux syscalls like `epoll`. Inspired by NGINX, it aims to provide a minimal yet powerful async architecture with configurable load balancing and dynamic request routing.

> ⚠️ This project is a work in progress (WIP). It's being built for learning, performance experimentation, and infrastructure tooling.

---

## ✨ Features (Planned)

- ✅ Low-level async I/O using `epoll` and `unix sockets`
- ✅ Reverse proxy support for HTTP protocols
- ✅ Pluggable load balancing (round-robin)
- ✅ Configurable via simple YAML files
- ⏳ Graceful start/stop/reload with signal handling
- ⏳ Prometheus-style metrics endpoint
