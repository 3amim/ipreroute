package ipreroute

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type IPReroute struct {
	next   http.Handler
	name   string
	redis  *redis.Client
	config *Config
	proxy  *httputil.ReverseProxy
}

type Config struct {
	RedisAddress string `json:"redis,omitempty"`
	RerouteKey   string `json:"rerouteKey,omitempty"`
	RerouteIP    string `json:"rerouteIP,omitempty"`
	ReroutePort  string `json:"reroutePort,omitempty"`
}

func CreateConfig() *Config {
	return &Config{
		RedisAddress: "redis:6379",
		RerouteKey:   "attacker_ip_",
		RerouteIP:    "127.0.0.1",
		ReroutePort:  "443",
	}
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {

	rdb := redis.NewClient(&redis.Options{
		Addr: config.RedisAddress,
	})

	serverName := config.RerouteIP
	if strings.Contains(serverName, ":") {
		serverName = strings.Split(serverName, ":")[0]
	}

	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
		},
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				newAddr := net.JoinHostPort(config.RerouteIP, config.ReroutePort)
				d := net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}
				return d.DialContext(ctx, network, newAddr)
			},
			MaxIdleConns:        500,
			MaxIdleConnsPerHost: 200,
			IdleConnTimeout:     90 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				ServerName:         serverName,
			},
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Println("Proxy error:", err)
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
		},
	}

	return &IPReroute{
		next:   next,
		name:   name,
		redis:  rdb,
		config: config,
		proxy:  proxy,
	}, nil
}

func (i *IPReroute) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	clientIP := getClientIP(req)
	key := i.config.RerouteKey + clientIP

	ctx, cancel := context.WithTimeout(req.Context(), 50*time.Millisecond)
	defer cancel()

	ipExists, err := i.redis.Exists(ctx, key).Result()
	if err != nil || ipExists != 1 {
		i.next.ServeHTTP(rw, req)
		return
	}

	ttl, err := i.redis.TTL(ctx, key).Result()
	if err != nil || ttl <= 0 {
		i.next.ServeHTTP(rw, req)
		return
	}

	log.Println("Silent HTTPS/HTTP reroute for IP:", clientIP, "TTL:", ttl)

	i.proxy.ServeHTTP(rw, req)
}

func getClientIP(req *http.Request) string {
	xff := req.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	ipPort := req.RemoteAddr
	ip := strings.Split(ipPort, ":")[0]
	return ip
}
