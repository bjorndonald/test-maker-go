package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bjorndonald/test-maker-service/constants"
	"github.com/bjorndonald/test-maker-service/database"
	"github.com/bjorndonald/test-maker-service/docs"
	"github.com/bjorndonald/test-maker-service/internal/bootstrap"
	"github.com/bjorndonald/test-maker-service/internal/helpers"
	"github.com/bjorndonald/test-maker-service/internal/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// @title Test Maker Service
// @version 1.0
// @description API documentation for Test Maker API
// @termsOfService http://swagger.io/terms/
// @contact.name Bjorn-Donald Bassey
// @contact.email bjorndonaldb@gmail.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host http://localhost:8000
// @BasePath /api/v1
func main() {
	g := gin.Default()

	docs.SwaggerInfo.BasePath = "/api/v1"
	constant := constants.New()

	flag.Parse()
	ctx := context.Background()

	g.Static("/assets", "./static/public")
	g.Static("/templates", "./templates")

	g.Use(gin.Logger())

	g.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// your custom format
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	g.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // add more origins
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	g.MaxMultipartMemory = 8 << 20

	g.GET("/api/v1/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	v1 := g.Group("/api/v1")

	dbConfig := database.Config{
		Host:     constant.DbHost,
		Port:     constant.DbPort,
		Password: constant.DbPassword,
		User:     constant.DbUser,
		DBName:   constant.DbName,
		SSLMode:  constant.SSLMode,
	}

	db, err := database.ConnectSQL(&dbConfig)
	if err != nil {
		log.Fatal(err)
	}

	dependencies := bootstrap.InitializeDependencies(db.SQL)

	routes.Routes(v1, dependencies)
	g.NoRoute(func(c *gin.Context) {
		helpers.ReturnError(c, "Something went wrong", fmt.Errorf("route not found"), http.StatusNotFound)
	})

	port := "8000"
	if port == "" {
		port = constant.Port
	}

	go log.Fatal(g.Run(":" + port))

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	sigusr1 := make(chan os.Signal, 1)
	signal.Notify(sigusr1, syscall.SIGUSR1)

	keepRunning := true

	for keepRunning {
		select {
		case <-ctx.Done():
			log.Println("terminating: context cancelled")
			keepRunning = false
		case <-sigterm:
			log.Println("terminating: via signal")
			keepRunning = false
		case <-sigusr1:
			// consumerClient.ToggleConsumptionFlow()
		}
	}
}
