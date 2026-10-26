package services

import (
	"strings"
	"sync"
	"time"
)

// Pembatas percobaan login (di memori, per proses server).
const (
	userFailLimit = 5                // gagal per kombinasi IP+username
	ipFailLimit   = 30               // gagal per IP untuk semua username
	failWindow    = 15 * time.Minute // jendela hitung
	lockDuration  = 15 * time.Minute // lama blokir sementara
)

type counter struct {
	n           int
	reset       time.Time
	lockedUntil time.Time
}

type limiter struct {
	mu sync.Mutex
	m  map[string]*counter
}

func newLimiter() *limiter { return &limiter{m: map[string]*counter{}} }

func (l *limiter) locked(key string, now time.Time) bool {
	c := l.m[key]
	return c != nil && now.Before(c.lockedUntil)
}

func (l *limiter) fail(key string, limit int, now time.Time) {
	c := l.m[key]
	if c == nil || now.After(c.reset) {
		c = &counter{reset: now.Add(failWindow)}
		l.m[key] = c
	}
	c.n++
	if c.n >= limit {
		c.lockedUntil = now.Add(lockDuration)
		c.n = 0
		c.reset = c.lockedUntil
	}
}

func (l *limiter) sweep(now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for k, c := range l.m {
		if now.After(c.reset) && now.After(c.lockedUntil) {
			delete(l.m, k)
		}
	}
}

var (
	loginUser = newLimiter()
	loginIP   = newLimiter()
)

func userKey(ip, username string) string { return ip + "|" + strings.ToLower(username) }

func LoginAllowed(ip, username string) bool {
	now := time.Now()
	loginUser.mu.Lock()
	u := loginUser.locked(userKey(ip, username), now)
	loginUser.mu.Unlock()
	loginIP.mu.Lock()
	i := loginIP.locked(ip, now)
	loginIP.mu.Unlock()
	return !u && !i
}

func LoginFailed(ip, username string) {
	now := time.Now()
	loginUser.mu.Lock()
	loginUser.fail(userKey(ip, username), userFailLimit, now)
	loginUser.mu.Unlock()
	loginIP.mu.Lock()
	loginIP.fail(ip, ipFailLimit, now)
	loginIP.mu.Unlock()
}

func LoginSucceeded(ip, username string) {
	loginUser.mu.Lock()
	delete(loginUser.m, userKey(ip, username))
	loginUser.mu.Unlock()
}

// RateLimiter membatasi jumlah request per kunci (IP) dalam satu jendela waktu.
type RateLimiter struct {
	l     *limiter
	limit int
	per   time.Duration
}

func NewRateLimiter(limit int, per time.Duration) *RateLimiter {
	r := &RateLimiter{l: newLimiter(), limit: limit, per: per}
	limiters = append(limiters, r.l)
	return r
}

func (r *RateLimiter) Allow(key string) bool {
	now := time.Now()
	r.l.mu.Lock()
	defer r.l.mu.Unlock()
	c := r.l.m[key]
	if c == nil || now.After(c.reset) {
		c = &counter{reset: now.Add(r.per)}
		r.l.m[key] = c
	}
	c.n++
	return c.n <= r.limit
}

var limiters = []*limiter{loginUser, loginIP}

func init() {
	go func() {
		for now := range time.Tick(5 * time.Minute) {
			for _, l := range limiters {
				l.sweep(now)
			}
		}
	}()
}
