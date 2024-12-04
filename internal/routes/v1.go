package routes

import (
	"github.com/bjorndonald/test-maker-service/internal/bootstrap"
	"github.com/gin-gonic/gin"
)

func Routes(r *gin.RouterGroup, d *bootstrap.AppDependencies) {
	RegisterRoutes(r, d)
}
