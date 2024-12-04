package middleware

import (
	"mime/multipart"
	"net/http"

	"github.com/bjorndonald/test-maker-service/internal/helpers"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func FileUploadMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			helpers.ReturnError(c, "File is required", err, http.StatusBadRequest)
			c.Abort()
			return
		}

		// Validate the file type
		if !isValidImage(file) {
			helpers.ReturnError(c, "Invalid file format", err, http.StatusBadRequest)
			c.Abort()
			return
		}

		fileName := "assets/documents/" + uuid.New().String() + file.Filename

		err = c.SaveUploadedFile(file, fileName)
		if err != nil {
			helpers.ReturnError(c, "File is required", err, http.StatusBadRequest)
			c.Abort()
			return
		}

		c.Set("file", fileName)
		c.Next()
	}
}

func isValidImage(header *multipart.FileHeader) bool {
	validTypes := map[string]bool{
		"application/pdf": true,
	}
	return validTypes[header.Header.Get("Content-Type")]
}
