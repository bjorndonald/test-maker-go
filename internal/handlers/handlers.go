package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bjorndonald/test-maker-service/internal/helpers"
	"github.com/bjorndonald/test-maker-service/internal/models"
	"github.com/bjorndonald/test-maker-service/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type Handler struct {
	docuRepo repository.DocumentInterface
}

func NewHandler(docuRepo repository.DocumentInterface) *Handler {
	return &Handler{
		docuRepo: docuRepo,
	}
}

type AnalyzedPDF struct {
	Id            string   `json:"id"`
	NumberOfPages int      `json:"numberOfPages"`
	Pdfs          []string `json:"pdfs"`
}

type LinkInput struct {
	Link string `json:"link" validate:"required"`
}

type Selection struct {
	From int `json:"from" validate:"required"`
	To   int `json:"to" validate:"required"`
}

type PagesInput struct {
	Id         string      `json:"id" validate:"required"`
	Selections []Selection `json:"selections" validate:"required"`
}

type QuestionInput struct {
	Id       string   `json:"id" validate:"required"`
	Num      int      `json:"num" validate:"required"`
	Subjects []string `json:"subjects" validate:"required"`
}

type AnalyzeResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    AnalyzedPDF `json:"data"`
}

type Question struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type QuestionResult struct {
	Questions []Question `json:"questions"`
}

type QuestionResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message"`
	Data    []Question `json:"data"`
}

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Success bool        `json:"success"`
}

// Analyze PDF
//
// @Summary Analyze pdf to retrieve pages
// @Description Analyze pdf to retrieve pages
// @Tags PDF
// @Accept json
// @Produce json
// @Success 200 {object} AnalyzeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /analyze [post]
func (a *Handler) AnalyzePdf(c *gin.Context) {
	filePath, exists := c.Get("file")
	if !exists {
		helpers.ReturnError(c, "File is required", errors.New("file is required"), http.StatusBadRequest)
		c.Abort()
		return
	}

	file, err := os.Open(filePath.(string))
	if !exists {
		helpers.ReturnError(c, "Issue reading file", err, http.StatusBadRequest)
		c.Abort()
		return
	}
	defer file.Close()

	numPages, err := api.PageCount(file, model.NewDefaultConfiguration())
	if err != nil {
		helpers.ReturnError(c, "Issue getting page number", err, http.StatusBadRequest)
		c.Abort()
		return
	}

	numWorkers := 5
	numTasks := numPages
	log.Println(numPages)
	jobs := make(chan int, numTasks)
	results := make(chan string, numTasks)

	var wg sync.WaitGroup

	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go func() {
			for pageNum := range jobs {
				encoded, err := helpers.ExtractPageAsBase64(filePath.(string), pageNum)
				if err != nil {
					helpers.ReturnError(c, "Issue reading file", err, http.StatusBadRequest)
					c.Abort()
					return
				}
				results <- encoded
			}
			wg.Done()
		}()
	}

	for i := 1; i <= numTasks; i++ {
		jobs <- i
		log.Printf("job: %d", i)
	}

	close(jobs)

	wg.Wait()
	close(results)

	pdfpages := []string{}

	for result := range results {
		pdfpages = append(pdfpages, result)
	}

	id, err := a.docuRepo.InsertDocument(c, models.Document{
		Id:        uuid.New(),
		Url:       filePath.(string),
		CreatedAt: time.Now(),
	})

	if err != nil {
		helpers.ReturnError(c, "Something went wrong", err, http.StatusInternalServerError)
		return
	}

	helpers.ReturnJSON(c, "Pdf analyzed succesfully", AnalyzedPDF{
		Id:            id,
		NumberOfPages: numPages,
		Pdfs:          pdfpages,
	}, http.StatusOK)
}

// Analyze PDF Link
//
// @Summary Analyze pdf link to retrieve pages
// @Description Analyze pdf link to retrieve pages
// @Tags PDF
// @Accept json
// @Produce json
// @Param credentials body LinkInput true "PDF Link"
// @Success 200 {object} AnalyzeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /analyze/link [post]
func (a *Handler) AnalyzeLink(c *gin.Context) {
	var input LinkInput
	validatedReqBody, exists := c.Get("validatedRequestBody")

	if !exists {
		helpers.ReturnError(c, "Something went wrong", fmt.Errorf(helpers.INVALID_REQUEST_BODY), http.StatusBadRequest)
		return
	}

	input, ok := validatedReqBody.(LinkInput)
	if !ok {
		helpers.ReturnError(c, "Something went wrong", fmt.Errorf(helpers.REQUEST_BODY_PARSE_ERROR), http.StatusBadRequest)
		return
	}

	newReq, err := http.NewRequest(http.MethodGet, input.Link, nil)

	client := &http.Client{}
	resp, err := client.Do(newReq)
	if err != nil {
		helpers.ReturnError(c, "Something went wrong", err, http.StatusInternalServerError)
		return
	}
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()
	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		helpers.ReturnError(c, "Something went wrong", err, http.StatusInternalServerError)
		return
	}

	inputFile := fmt.Sprintf("assets/documents/%s.pdf", uuid.NewString())
	err = os.WriteFile(inputFile, data, os.ModeAppend)
	if err != nil {
		helpers.ReturnError(c, "Something went wrong", err, http.StatusInternalServerError)
		return
	}

	file, err := os.Open(inputFile)
	if !exists {
		helpers.ReturnError(c, "Issue reading file", err, http.StatusBadRequest)
		c.Abort()
		return
	}
	defer file.Close()

	numPages, err := api.PageCount(file, model.NewDefaultConfiguration())
	if err != nil {
		helpers.ReturnError(c, "Issue getting page number", err, http.StatusBadRequest)
		c.Abort()
		return
	}

	numWorkers := 5
	numTasks := numPages
	log.Println(numPages)
	jobs := make(chan int, numTasks)
	results := make(chan string, numTasks)

	var wg sync.WaitGroup

	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go func() {
			for pageNum := range jobs {
				encoded, err := helpers.ExtractPageAsBase64(inputFile, pageNum)
				if err != nil {
					helpers.ReturnError(c, "Issue reading file", err, http.StatusBadRequest)
					c.Abort()
					return
				}
				results <- encoded
			}
			wg.Done()
		}()
	}

	for i := 1; i <= numTasks; i++ {
		jobs <- i
		log.Printf("job: %d", i)
	}

	close(jobs)

	wg.Wait()
	close(results)

	pdfpages := []string{}

	for result := range results {
		pdfpages = append(pdfpages, result)
	}

	id, err := a.docuRepo.InsertDocument(c, models.Document{
		Id:        uuid.New(),
		Url:       inputFile,
		CreatedAt: time.Now(),
	})

	helpers.ReturnJSON(c, "Pdf analyzed succesfully", AnalyzedPDF{
		Id:            id,
		NumberOfPages: numPages,
		Pdfs:          pdfpages,
	}, http.StatusOK)
}

// Embed pages of the pdf
//
// @Summary Embed PDF
// @Description Embed PDF
// @Tags PDF
// @Accept json
// @Produce json
// @Param credentials body PagesInput true "PDF pages"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /embed [post]
func (a *Handler) EmbedPages(c *gin.Context) {
	var pages PagesInput
	validatedReqBody, exists := c.Get("validatedRequestBody")

	if !exists {
		helpers.ReturnError(c, "Something went wrong", fmt.Errorf(helpers.INVALID_REQUEST_BODY), http.StatusBadRequest)
		return
	}

	pages, ok := validatedReqBody.(PagesInput)
	if !ok {
		helpers.ReturnError(c, "Something went wrong", fmt.Errorf(helpers.REQUEST_BODY_PARSE_ERROR), http.StatusBadRequest)
		return
	}

	documentId, err := uuid.Parse(pages.Id)
	if err != nil {
		helpers.ReturnError(c, "Error parsing document ID", err, http.StatusInternalServerError)
		c.Abort()
		return
	}

	doc, err := a.docuRepo.RetrieveDocument(c, pages.Id)
	if err != nil {
		helpers.ReturnError(c, "Issue assessing database", err, http.StatusInternalServerError)
		c.Abort()
		return
	}

	selectedPages := []int{}
	for _, selection := range pages.Selections {
		for i := selection.From; i <= selection.To; i++ {
			selectedPages = append(selectedPages, i)
		}
	}

	text, err := helpers.ExtractPDFText(doc.Url, selectedPages)
	if err != nil {
		helpers.ReturnError(c, "Text extraction error", err, http.StatusInternalServerError)
		c.Abort()
		return
	}

	sentences := helpers.TokenizeSentences(text)

	embeddings, err := helpers.GetOpenAIEmbeddings(sentences)
	if err != nil {
		helpers.ReturnError(c, "Embedding error", err, http.StatusInternalServerError)
		c.Abort()
		return
	}

	chunks := []models.Chunk{}
	log.Println(len(sentences))
	log.Println(len(embeddings))
	for i, embedding := range embeddings {
		chunks = append(chunks, models.Chunk{
			Id:             uuid.New(),
			DocumentId:     documentId,
			Chunk:          sentences[i],
			ChunkEmbedding: embedding,
		})
	}

	err = a.docuRepo.InsertChunks(c, chunks)
	if err != nil {
		helpers.ReturnError(c, "Embedding error", err, http.StatusInternalServerError)
		c.Abort()
		return
	}

	helpers.ReturnJSON(c, "Pages embedded succesfully", nil, http.StatusOK)
}

// Generate questions for the PDF
//
// @Summary Generate Questions
// @Description GenerateQuestions
// @Tags PDF
// @Accept json
// @Produce json
// // @Param credentials body QuestionInput true "PDF pages"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /generate [post]
func (a *Handler) GenerateQuestions(c *gin.Context) {
	var question QuestionInput
	validatedReqBody, exists := c.Get("validatedRequestBody")

	if !exists {
		helpers.ReturnError(c, "Something went wrong", fmt.Errorf(helpers.INVALID_REQUEST_BODY), http.StatusBadRequest)
		return
	}

	question, ok := validatedReqBody.(QuestionInput)
	if !ok {
		helpers.ReturnError(c, "Something went wrong", fmt.Errorf(helpers.REQUEST_BODY_PARSE_ERROR), http.StatusBadRequest)
		return
	}

	prompt := ""
	if len(question.Subjects) > 0 {
		prompt = fmt.Sprintf("Please generate a list of %d questions: %s", question.Num, strings.Join(question.Subjects, ", "))
	} else {
		prompt = fmt.Sprintf("Please generate a list of %d  questions", question.Num)
	}

	embeds, err := helpers.GetOpenAIEmbeddings([]string{prompt})
	if err != nil {
		helpers.ReturnError(c, "Embedding error", err, http.StatusInternalServerError)
		c.Abort()
		return
	}

	chunks, err := a.docuRepo.VectorSearch(c, question.Id, embeds[0])
	if err != nil {
		helpers.ReturnError(c, "Search error", err, http.StatusInternalServerError)
		c.Abort()
		return
	}

	context := ""
	for _, v := range chunks {
		context += v.Chunk + "\n"
	}

	choices, err := helpers.GenerateQuestions(c, prompt, context)
	if err != nil {
		helpers.ReturnError(c, "Generating error", err, http.StatusInternalServerError)
		c.Abort()
		return
	}

	var result QuestionResult

	log.Println(choices[0].Message.Content)
	questionRes := strings.ReplaceAll(strings.ReplaceAll(choices[0].Message.Content, "```", ""), "json", "")
	questionRes = strings.ReplaceAll(questionRes, `'`, `"`)
	err = json.Unmarshal([]byte(questionRes), &result)
	if err != nil {
		helpers.ReturnError(c, "Result parsing error", err, http.StatusInternalServerError)
		c.Abort()
		return
	}

	helpers.ReturnJSON(c, "Questions retrieved succesfully", result.Questions, http.StatusOK)
}
