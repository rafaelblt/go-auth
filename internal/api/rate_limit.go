package api

import (
	"math"
	"net/http"
	"net/netip"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/rafaelblt/go-auth/internal/port"
)

type endpointRateLimit struct {
	limiter        port.RateLimiter
	trustedProxies []netip.Prefix
	endpoint       string
	limit          port.RateLimit
}

func newEndpointRateLimit(cfg RateLimiting, endpoint string, limit port.RateLimit) *endpointRateLimit {
	rateLimit := endpointRateLimit{
		limiter:        cfg.Limiter,
		trustedProxies: slices.Clone(cfg.TrustedProxies),
		endpoint:       endpoint,
		limit:          limit,
	}
	return &rateLimit
}

func validRateLimit(limit port.RateLimit) bool {
	return limit.Requests > 0 && limit.Period > 0
}

func (l *endpointRateLimit) allow(r *http.Request) (decision port.RateLimitDecision, client string, err error) {
	client = clientAddress(r, l.trustedProxies)
	decision, err = l.limiter.Allow(r.Context(), l.endpoint+" "+client, l.limit)
	return decision, client, err
}

// See docs/architecture/http.md#client-address.
func clientAddress(r *http.Request, trustedProxies []netip.Prefix) string {
	peer, err := netip.ParseAddrPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	client := normalizeAddr(peer.Addr())
	forwardedFor := strings.Split(strings.Join(r.Header.Values("X-Forwarded-For"), ","), ",")
	for i := len(forwardedFor) - 1; i >= 0 && isTrustedProxy(client, trustedProxies); i-- {
		addr, err := netip.ParseAddr(strings.TrimSpace(forwardedFor[i]))
		if err != nil {
			break
		}
		client = normalizeAddr(addr)
	}

	if client.Is4() {
		return client.String()
	}
	prefix, _ := client.Prefix(64)
	return prefix.String()
}

func normalizeAddr(addr netip.Addr) netip.Addr {
	return addr.Unmap().WithZone("")
}

func isTrustedProxy(addr netip.Addr, trustedProxies []netip.Prefix) bool {
	return slices.ContainsFunc(trustedProxies, func(prefix netip.Prefix) bool {
		return prefix.Contains(addr)
	})
}

func retryAfterSeconds(d time.Duration) string {
	seconds := int(math.Ceil(d.Seconds()))
	return strconv.Itoa(max(seconds, 1))
}
