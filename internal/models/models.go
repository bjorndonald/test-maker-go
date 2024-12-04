package models

import (
	"time"

	"github.com/google/uuid"
)

type Chunk struct {
	Id             uuid.UUID
	DocumentId     uuid.UUID
	Chunk          string
	ChunkEmbedding []float32
}

type Document struct {
	Id        uuid.UUID
	Url       string
	CreatedAt time.Time
}
