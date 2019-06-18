package route

import (
	"github.com/gin-gonic/gin"
	"somethings-funny/service/route/log"
	"somethings-funny/service/route/register"
	"somethings-funny/service/route/upload"
)

func NewRouter(piDomain, hostedZone, region, bucket *string) *gin.Engine {
	router := gin.New()
	router.Use(
		gin.LoggerWithWriter(log.NewInfoWriter()),
		gin.RecoveryWithWriter(log.NewErrorWriter()),
	)
	router.Handle("POST", "/register-pi", register.HandleRegisterPi(piDomain, hostedZone))
	router.Handle("POST", "/upload", upload.UploadHandler(region, bucket))
	return router
}
