
# IP Reroute Middleware for Traefik

A lightweight, fast, and **production-ready middleware** for **transparent rerouting** of client requests based on IP using Redis.

This middleware allows you to reroute requests to a different backend **without using redirects (302/301)** and without modifying request headers, so the client remains unaware of the reroute.

---

## ✨ Features

- ✅ Fully transparent reroute (no redirects)
- ✅ Supports both HTTP and HTTPS
- ✅ Preserves all headers, including `Host`
- ✅ HTTPS SNI support
- ✅ Direct Redis integration for decision-making
- ✅ High-load ready (handles thousands of RPS)
- ✅ No internal cache (real-time decisions from Redis)
- ✅ Safe Redis timeout to prevent blocking
- ✅ Reused connection pool and transport for performance
- ✅ Ready to use as a Traefik plugin

---

## 🧠 How It Works

For each incoming request:

1. Extracts the client IP (from `X-Forwarded-For` or `RemoteAddr`)
2. Checks Redis for a key corresponding to the IP
3. If the key exists and is not expired:
   - The request is routed **silently** to the configured target IP/port
   - The original headers and host are preserved
4. If the key does not exist or has expired:
   - The request proceeds normally to the original backend

---

## ⚙️ Configuration

```go
type Config struct {
    RedisAddress string // e.g., "redis:6379"
    RerouteKey   string // e.g., "attacker_ip_"
    RerouteIP    string // target IP for reroute
    ReroutePort  string // target port
}
```

Example default configuration:

```go
&Config{
    RedisAddress: "redis:6379",
    RerouteKey:   "attacker_ip_",
    RerouteIP:    "127.0.0.1",
    ReroutePort:  "443",
}
```

---

## ⚡ Performance Notes

* Reuses a single `httputil.ReverseProxy` and `http.Transport` for all rerouted requests
* MaxIdleConns and MaxIdleConnsPerHost are tuned for high concurrency
* Redis timeout ensures requests do not block if Redis is slow
* Suitable for several thousand concurrent requests per second on modern servers

---

## 🛠 Usage

Integrate this middleware into Traefik or any Go HTTP stack:

```go
middleware, err := ipreroute.New(ctx, nextHandler, config, "ip_reroute")
if err != nil {
    log.Fatal(err)
}

http.Handle("/", middleware)
```

---

## ⚠️ Notes

* This middleware **does not implement any caching**. All decisions are made in real-time from Redis.
* TLS verification is skipped (`InsecureSkipVerify: true`) for maximum transparency and simplicity. Ensure this fits your security requirements.
* Make sure your Redis instance can handle high request rates if deployed at scale.

---

## 📝 License

MIT License

```md
If you want, I can also write a **shorter, GitHub-friendly description with badges, setup instructions, and example Docker/Trafik usage** to make it look professional on the repo homepage.  

Do you want me to do that too?
```
