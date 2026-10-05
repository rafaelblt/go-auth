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
	if isTrustedProxy(client, trustedProxies) {
		client = forwardedClient(r.Header.Values("X-Forwarded-For"), client, trustedProxies)
	}

	if client.Is4() {
		return client.String()
	}
	prefix, _ := client.Prefix(64)
	return prefix.String()
}

// forwardedClient cuts the header from the right, one entry at a time, rather
// than splitting it: its length is the client's choice, and the walk stops at
// the first address that is not a trusted proxy.
func forwardedClient(header []string, peer netip.Addr, trustedProxies []netip.Prefix) netip.Addr {
	client := peer
	forwardedFor := strings.Join(header, ",")
	for isTrustedProxy(client, trustedProxies) {
		comma := strings.LastIndexByte(forwardedFor, ',')
		addr, err := netip.ParseAddr(strings.TrimSpace(forwardedFor[comma+1:]))
		if err != nil {
			return client
		}
		client = normalizeAddr(addr)
		if comma < 0 {
			return client
		}
		forwardedFor = forwardedFor[:comma]
	}
	return client
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
