-- Habilitar la extensión pgvector en PostgreSQL
CREATE EXTENSION IF NOT EXISTS vector;

-- Añadir la columna de embeddings a la tabla de canciones con tamaño de 64 dimensiones
ALTER TABLE songs ADD COLUMN IF NOT EXISTS embedding vector(64);

-- Crear un índice HNSW para búsquedas ultrarrápidas de similitud de coseno
CREATE INDEX IF NOT EXISTS songs_embedding_hnsw_idx ON songs USING hnsw (embedding vector_cosine_ops);
