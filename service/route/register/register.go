package register

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net"
	"somethings-funny/service/dns"
)

func HandleRegisterPi(piDomain, hostedZone *string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip, ok := c.GetPostForm("ip")
		if !ok {
			_ = c.AbortWithError(400, errors.New("missing required field: ip"))
			return
		}
		if parsed := net.ParseIP(ip); parsed == nil {
			_ = c.AbortWithError(400, errors.New("malformed ip"))
			return
		}
		err := dns.RegisterIP(c, piDomain, &ip, hostedZone)
		if err != nil {
			_ = c.AbortWithError(500, err)
			return
		}
		c.AbortWithStatus(200)
	}
}
