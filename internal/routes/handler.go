package routes

import (
	"github.com/bjorndonald/test-maker-service/internal/bootstrap"
	"github.com/bjorndonald/test-maker-service/internal/handlers"
	"github.com/bjorndonald/test-maker-service/internal/middleware"
	"github.com/bjorndonald/test-maker-service/internal/repository"
	"github.com/bjorndonald/test-maker-service/internal/validators"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, d *bootstrap.AppDependencies) {
	repo := repository.NewPostgresRepo(d.DatabaseService)
	handler := handlers.NewHandler(repo)
	router.POST("/analyze", middleware.FileUploadMiddleware(), handler.AnalyzePdf)
	router.POST("/analyze/link", validators.ValidateLinkSchema, handler.AnalyzeLink)
	router.POST("/embed", validators.ValidatePagesSchema, handler.EmbedPages)
	router.POST("/generate", validators.ValidateQuestionSchema, handler.GenerateQuestions)
}
