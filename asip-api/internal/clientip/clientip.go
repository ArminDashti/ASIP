package clientip

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// FromContext returns the caller IP. When Cloudflare terminates TLS it sets
// CF-Connecting-IP; prefer that over Gin's X-Forwarded-For parsing (which can
// surface a Cloudflare edge address instead of the browser).
func FromContext(c *gin.Context) string {
	if cf := strings.TrimSpace(c.GetHeader("CF-Connecting-IP")); cf != "" {
		if parsed := net.ParseIP(cf); parsed != nil {
			return parsed.String()
		}
	}
	return c.ClientIP()
}
