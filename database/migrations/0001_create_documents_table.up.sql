CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE chunks (
    id UUID NOT NULL,
    document UUID NOT NULL,
    chunk TEXT NOT NULL,
    chunk_embedding VECTOR(1536)
);

CREATE INDEX ON chunks USING ivfflat (chunk_embedding vector_cosine_ops) WITH (lists = 100);

CREATE TABLE documents (
    id UUID NOT NULL,
    url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL
);