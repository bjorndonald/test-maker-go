package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/bjorndonald/test-maker-service/internal/models"
)

type DocumentInterface interface {
	InsertChunks(ctx context.Context, chunks []models.Chunk) error
	InsertDocument(ctx context.Context, doc models.Document) (string, error)
	RetrieveDocument(ctx context.Context, id string) (models.Document, error)
	VectorSearch(ctx context.Context, document_id string, prompt []float32) ([]models.Chunk, error)
}

type documentRepo struct {
	DB *sql.DB
}

func NewPostgresRepo(conn *sql.DB) DocumentInterface {
	return &documentRepo{
		DB: conn,
	}
}

func (m *documentRepo) RetrieveDocument(ctx context.Context, id string) (models.Document, error) {
	var document models.Document

	query := `
		select id, url, created_at from documents where id = $1
	`

	row := m.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&document.Id,
		&document.Url,
		&document.CreatedAt,
	)
	if err != nil {
		return document, err
	}

	return document, nil
}

func (m *documentRepo) VectorSearch(ctx context.Context, document_id string, prompt []float32) ([]models.Chunk, error) {
	var chunks []models.Chunk

	query := `
		SELECT id, chunk, chunk_embedding <#> $1 AS similarity
		FROM chunks WHERE document = $2
		ORDER BY similarity
		LIMIT 5;
	`
	vector := fmt.Sprintf("[%s]", strings.Trim(strings.Replace(fmt.Sprint(prompt), " ", ",", -1), "[]"))

	rows, err := m.DB.Query(query, vector, document_id)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var chunk models.Chunk
		err := rows.Scan(&chunk.Id, &chunk.Chunk, &chunk.ChunkEmbedding)
		if err != nil {
			panic(err)
		}
	}

	return chunks, nil
}

func (m *documentRepo) InsertDocument(ctx context.Context, doc models.Document) (string, error) {
	var newID string
	stmt := `
		insert into documents (id, url, created_at) values ($1, $2, $3) returning id 
		`
	err := m.DB.QueryRowContext(ctx, stmt,
		doc.Id,
		doc.Url,
		doc.CreatedAt,
	).Scan(&newID)
	if err != nil {
		return "", err
	}
	log.Println(newID)

	return newID, nil
}

func (m *documentRepo) InsertChunks(ctx context.Context, chunks []models.Chunk) error {
	for _, chunk := range chunks {
		var newID string
		stmt := `
			insert into chunks (id, document, chunk, chunk_embedding) values ($1, $2, $3, $4::vector) returning id 
		`
		vector := fmt.Sprintf("[%s]", strings.Trim(strings.Replace(fmt.Sprint(chunk.ChunkEmbedding), " ", ",", -1), "[]"))
		// log.Printf("Query: vector=%s\n", vector)
		err := m.DB.QueryRowContext(ctx, stmt,
			chunk.Id,
			chunk.DocumentId,
			chunk.Chunk,
			vector,
		).Scan(&newID)
		if err != nil {
			return err
		}
	}

	return nil
}
