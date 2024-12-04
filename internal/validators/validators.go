package validators

import (
	"net/http"

	"github.com/bjorndonald/test-maker-service/internal/handlers"
	"github.com/bjorndonald/test-maker-service/internal/helpers"
	"github.com/bjorndonald/test-maker-service/validator"
	"github.com/gin-gonic/gin"
)

func ValidateLinkSchema(c *gin.Context) {
	var body handlers.LinkInput
	bindAndValidate(c, &body)
	c.Set("validatedRequestBody", body)
	c.Next()
}

func ValidatePagesSchema(c *gin.Context) {
	var body handlers.PagesInput
	bindAndValidate(c, &body)
	c.Set("validatedRequestBody", body)
	c.Next()
}

func ValidateQuestionSchema(c *gin.Context) {
	var body handlers.QuestionInput
	bindAndValidate(c, &body)
	c.Set("validatedRequestBody", body)
	c.Next()
}

func bindAndValidate(c *gin.Context, body interface{}) {
	if err := c.ShouldBindJSON(body); err != nil {
		helpers.ReturnError(c, "Error validating input", err, http.StatusBadRequest)
		c.Abort()
		return
	}

	if err := validator.Validate(body); err != nil {
		helpers.ReturnError(c, "Error validating input", err, http.StatusBadRequest)
		c.Abort()
		return
	}
}
