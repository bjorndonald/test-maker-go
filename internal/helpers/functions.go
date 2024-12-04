package helpers

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/bjorndonald/test-maker-service/constants"
	"github.com/dslipak/pdf"
	"github.com/gin-gonic/gin"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/sashabaranov/go-openai"
)

func ReturnJSON(c *gin.Context, message string, data interface{}, statusCode int) {
	c.Status(statusCode)
	c.JSON(statusCode, gin.H{
		"status":  statusCode <= 201,
		"message": message,
		"data":    data,
	})
}

func ReturnError(c *gin.Context, message string, err error, status int) {
	c.JSON(status, gin.H{
		"message": message,
		"data":    err.Error(),
		"status":  false,
	})
	log.Println("error: ", err.Error())
	log.Println("message: ", message)
}

func ExtractPageAsBase64(inputPDF string, pageNum int) (string, error) {
	// Open the PDF file
	file, err := os.Open(inputPDF)
	if err != nil {
		return "", fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	// Create a temporary output directory for the extracted page
	outputDir := "./assets/documents"
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("could not create output directory: %w", err)
	}

	err = api.ExtractPages(file, outputDir, "", []string{fmt.Sprintf("%d", pageNum)}, model.NewDefaultConfiguration())
	if err != nil {
		return "", fmt.Errorf("failed to extract page %d: %w", pageNum, err)
	}

	// Read the extracted page file
	extractedPagePath := fmt.Sprintf("%s/._page_%d.pdf", outputDir, pageNum)
	extractedFile, err := os.Open(extractedPagePath)
	if err != nil {
		return "", fmt.Errorf("could not open extracted page file: %w", err)
	}
	defer extractedFile.Close()

	// Read the file content into a buffer
	buffer := new(bytes.Buffer)
	_, err = io.Copy(buffer, extractedFile)
	if err != nil {
		return "", fmt.Errorf("could not read extracted page content: %w", err)
	}

	err = os.Remove(extractedPagePath)
	if err != nil {
		return "", fmt.Errorf("could not read extracted page content: %w", err)
	}

	// Encode the buffer content to Base64
	base64String := base64.StdEncoding.EncodeToString(buffer.Bytes())

	return base64String, nil
}

func ExtractPDFText(pdfPath string, selectedPages []int) (string, error) {
	reader, err := pdf.Open(pdfPath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %v", err)
	}

	var fullText strings.Builder
	for _, pageIndex := range selectedPages {
		page := reader.Page(pageIndex)

		// Extract text from the page
		content := page.Content()

		for _, text := range content.Text {
			fullText.WriteString(text.S)
		}

		fullText.WriteString("\n")
	}

	return fullText.String(), nil
}

// tokenizeSentences splits text into sentences using a simple approach
func TokenizeSentences(text string) []string {
	// Replace common abbreviations to prevent incorrect sentence splitting
	text = strings.ReplaceAll(text, "Mr.", "Mr")
	text = strings.ReplaceAll(text, "Mrs.", "Mrs")
	text = strings.ReplaceAll(text, "Dr.", "Dr")

	// Split text into sentences
	sentences := strings.FieldsFunc(text, func(r rune) bool {
		return r == '.' || r == '!' || r == '?'
	})

	// Clean and trim sentences
	var cleanSentences []string
	for _, sentence := range sentences {
		trimmed := strings.TrimSpace(sentence)
		if trimmed != "" {
			// Restore punctuation
			if strings.Contains(text, trimmed+".") {
				trimmed += "."
			} else if strings.Contains(text, trimmed+"!") {
				trimmed += "!"
			} else if strings.Contains(text, trimmed+"?") {
				trimmed += "?"
			}
			cleanSentences = append(cleanSentences, trimmed)
		}
	}

	return cleanSentences
}

func GetOpenAIEmbeddings(texts []string) ([][]float32, error) {

	constant := constants.New()

	// Prepare request payload
	client := openai.NewClient(constant.OpenAIKey)

	resp, err := client.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
		Model: openai.AdaEmbeddingV2,
		Input: texts,
	})

	if err != nil {
		return [][]float32{}, err
	}

	embeddings := [][]float32{}

	for _, d := range resp.Data {
		embeddings = append(embeddings, d.Embedding)
	}

	return embeddings, nil
}

func GenerateQuestions(ctx context.Context, prompt string, context string) ([]openai.ChatCompletionChoice, error) {
	constant := constants.New()

	client := openai.NewClient(constant.OpenAIKey)

	systemPrompt := fmt.Sprintf(RESPONSE_SYSTEM_TEMPLATE, "", FORMAT_INSTRUCTIONS)

	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return []openai.ChatCompletionChoice{}, err
	}

	return resp.Choices, nil
}
