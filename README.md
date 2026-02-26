# phone.seek()

Multimodal semantic search engine for smartphones, powered by vector embeddings and Qdrant.

Search phones by **natural language** ("good camera phone under 500") or by **uploading a photo** of a phone you like. Supports both text and image queries with real-time filtering.

## Architecture

```
Frontend (Vue 3 + DaisyUI)     :5173
    |
    | /api proxy
    v
Backend (Go)                   :8080 (internal)
    |           |
    | gRPC      | HTTP
    v           v
Qdrant DB    Embedder (FastAPI)
  :6333        :8000 (internal)
```

**4 services**, fully dockerized:

| Service | Stack | Role |
|---------|-------|------|
| **frontend** | Vue 3, Tailwind CSS 4, DaisyUI 5 | Search UI with text/image/camera input and filters |
| **backend** | Go 1.26, qdrant-go-client | API server, CSV parsing, data seeding, image proxy |
| **embedder** | Python 3.12, FastAPI, PyTorch | Generates vector embeddings from text and images |
| **qdrant** | Qdrant (latest) | Vector database with cosine similarity search |

## Embedding Models

| Model | Purpose | Dimensions | Quantization |
|-------|---------|------------|--------------|
| [BAAI/bge-m3](https://huggingface.co/BAAI/bge-m3) | Text embeddings | 1024 | INT8 (PyTorch dynamic) |
| [CLIP ViT-B/32](https://huggingface.co/sentence-transformers/clip-ViT-B-32) | Image embeddings | 512 | INT8 (PyTorch dynamic) |

**BGE-M3** is multilingual (100+ languages) with an 8192 token context window, so you can search in any language against the English phone specs data.

Both models are quantized to INT8 at startup for faster CPU inference.

## Data Pipeline

1. **Parse** ~10,000 smartphones from the GSMArena CSV dataset (40+ specs per phone)
2. **Download** phone images concurrently (10 workers)
3. **Embed** text descriptions with BGE-M3 (1024d vectors) in batches
4. **Embed** images with CLIP (512d vectors) in batches
5. **Store** in Qdrant as named vectors (`text` + `image`) with full payload
6. **Index** payload fields for filtering (brand, OS, display type, NFC, network, price)

The seeding runs automatically on first startup if the collection doesn't exist.

## Search Features

- **Text search**: natural language queries, multilingual
- **Image search**: upload a photo or use your camera
- **Filters**: brand, OS family, display type, NFC, network technology, price range (EUR)
- **Cosine similarity score** displayed on each result card

## Quick Start

```bash
# Clone and configure
cp .env.example .env  # adjust UID/GID if needed

# Start all services
docker compose up -d --build

# First run takes a while (model download + data seeding)
# Watch progress:
docker compose logs -f embedder backend

# Open the UI
open http://localhost:5173
```

## Environment Variables

Create a `.env` file:

```env
UID=1000
GID=1000
```

## Project Structure

```
.
├── backend/                 # Go API server
│   ├── cmd/server/          # Entry point
│   └── internal/
│       ├── model/           # Smartphone domain model
│       ├── csvparser/       # CSV parsing
│       ├── qdrant/          # Seeder + Searcher
│       ├── embedder/        # HTTP client for embedder
│       └── server/          # HTTP handlers
├── embedder/                # Python embedding service
│   └── main.py              # FastAPI + CLIP + BGE-M3
├── frontend/                # Vue 3 SPA
│   └── src/components/      # SearchBar, SearchFilters, PhoneCard
├── qdrant/                  # Qdrant container config
├── data/                    # smartphones.csv (not in repo)
├── images/                  # Downloaded phone images (not in repo)
├── models_cache/            # Cached ML models (not in repo)
└── docker-compose.yml
```

## Tech Decisions

- **Named vectors** in Qdrant allow text and image search on the same collection without duplicating data
- **BGE-M3** over MiniLM for multilingual support and 8192 token context (fits all phone specs without truncation)
- **INT8 quantization** for CPU inference performance (PyTorch dynamic quantization on Linear layers)
- **Go backend** handles concurrent image downloads and batch embedding requests efficiently
- **Glassmorphism UI** with semi-transparent cards and backdrop blur

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/search?q=...` | Text search with optional filters |
| POST | `/api/search/image` | Image search (multipart form) |
| GET | `/api/filters` | Available filter options |
| GET | `/api/images/:file` | Serve phone images |
| GET | `/health` | Health check |
