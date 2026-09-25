package hubserver

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pagbest154-cmd/system-monitor/internal/fleet"
)

const sessionCookie = "sm_hub_session"
const sessionTTL = 7 * 24 * 3600

var agentAPIPathRE = regexp.MustCompile(`^/api/agents/([^/]+)/(.+)$`)

func hubCredentials() (string, string, bool) {
	name := strings.TrimSpace(os.Getenv("HUB_NAME"))
	key := strings.TrimSpace(os.Getenv("HUB_KEY"))
	if name != "" && key != "" {
		return name, key, true
	}
	return "", "", false
}

func HubAuthEnabled() bool {
	_, _, ok := hubCredentials()
	return ok
}

func HubName() string {
	name, _, ok := hubCredentials()
	if ok {
		return name
	}
	return ""
}

func VerifyCredentials(name, key string) bool {
	hubName, hubKey, ok := hubCredentials()
	if !ok {
		return true
	}
	return subtle.ConstantTimeCompare([]byte(name), []byte(hubName)) == 1 &&
		subtle.ConstantTimeCompare([]byte(key), []byte(hubKey)) == 1
}

func sessionSecret() []byte {
	_, key, ok := hubCredentials()
	material := key
	if !ok {
		material = "system-monitor-dev"
	}
	sum := sha256.Sum256([]byte(material))
	return sum[:]
}

func CreateSessionToken() string {
	expires := time.Now().Unix() + sessionTTL
	nonce := randomToken(16)
	payload := strconv.FormatInt(expires, 10) + "." + nonce
	sig := hmacSHA256(payload)
	raw := payload + "." + sig
	return strings.TrimRight(base64.URLEncoding.EncodeToString([]byte(raw)), "=")
}

func VerifySessionToken(token string) bool {
	if token == "" {
		return false
	}
	padding := strings.Repeat("=", (4-len(token)%4)%4)
	raw, err := base64.URLEncoding.DecodeString(token + padding)
	if err != nil {
		return false
	}
	parts := strings.Split(string(raw), ".")
	if len(parts) != 3 {
		return false
	}
	payload := parts[0] + "." + parts[1]
	if hmacSHA256(payload) != parts[2] {
		return false
	}
	expires, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return false
	}
	return expires > time.Now().Unix()
}

func hmacSHA256(payload string) string {
	mac := hmac.New(sha256.New, sessionSecret())
	mac.Write([]byte(payload))
	return hexEncode(mac.Sum(nil))
}

func hexEncode(b []byte) string {
	const hexdigits = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, v := range b {
		out[i*2] = hexdigits[v>>4]
		out[i*2+1] = hexdigits[v&0x0f]
	}
	return string(out)
}

func randomToken(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}

func ParseBasicAuth(header string) (string, string, bool) {
	if !strings.HasPrefix(strings.ToLower(header), "basic ") {
		return "", "", false
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(header[6:]))
	if err != nil {
		return "", "", false
	}
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func IsAuthenticated(r *http.Request) bool {
	if !HubAuthEnabled() {
		return true
	}
	if token, err := r.Cookie(sessionCookie); err == nil && VerifySessionToken(token.Value) {
		return true
	}
	if user, pass, ok := ParseBasicAuth(r.Header.Get("Authorization")); ok && VerifyCredentials(user, pass) {
		return true
	}
	return false
}

func agentPathInfo(path, method string) (agentID string, ok bool) {
	m := agentAPIPathRE.FindStringSubmatch(path)
	if len(m) != 3 {
		return "", false
	}
	switch method {
	case http.MethodPost:
		switch m[2] {
		case "metrics", "heartbeat", "config":
			return m[1], true
		}
	case http.MethodGet:
		if m[2] == "notify" || strings.HasPrefix(m[2], "notify/") {
			return m[1], true
		}
	}
	return "", false
}

func IsAgentAuthenticated(r *http.Request) bool {
	agentID, ok := agentPathInfo(r.URL.Path, r.Method)
	if !ok {
		return false
	}
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		return false
	}
	return fleet.VerifyAgentToken(agentID, strings.TrimSpace(auth[7:]))
}

func IsPublicPath(path, method string) bool {
	if strings.HasPrefix(path, "/static/") {
		return true
	}
	if path == "/login" {
		return true
	}
	if path == "/api/auth/login" && method == http.MethodPost {
		return true
	}
	if path == "/api/auth/status" && method == http.MethodGet {
		return true
	}
	if path == "/api/version" && method == http.MethodGet {
		return true
	}
	return false
}

func SessionCookieHeader(token string, secure bool) string {
	parts := []string{
		sessionCookie + "=" + token,
		"Path=/", "HttpOnly", "SameSite=Lax",
		"Max-Age=" + strconv.Itoa(sessionTTL),
	}
	if secure {
		parts = append(parts, "Secure")
	}
	return strings.Join(parts, "; ")
}

func ClearSessionCookieHeader(secure bool) string {
	parts := []string{sessionCookie + "=", "Path=/", "HttpOnly", "SameSite=Lax", "Max-Age=0"}
	if secure {
		parts = append(parts, "Secure")
	}
	return strings.Join(parts, "; ")
}

func RequestIsSecure(r *http.Request) bool {
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		return strings.EqualFold(strings.TrimSpace(strings.Split(proto, ",")[0]), "https")
	}
	return r.TLS != nil
}
