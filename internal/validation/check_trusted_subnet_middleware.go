package validation

import (
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CheckTrustedSubnetMiddleware(trustedSubnet string) gin.HandlerFunc {
	var ipnet *net.IPNet
	if trustedSubnet != "" {
		_, network, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			log.Fatalf("invalid CIDR in trusted_subnet: %v", err)
		}
		ipnet = network
	}

	return func(c *gin.Context) {
		if ipnet != nil {
			realIP := c.GetHeader("X-Real-IP")
			ip := net.ParseIP(realIP)
			if ip == nil || !ipnet.Contains(ip) {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}
		c.Next()
	}
}
