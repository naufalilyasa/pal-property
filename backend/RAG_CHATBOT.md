# RAG Chatbot Backend API

## Scope v1
- Backend API only
- Property-only answers and recommendations
- Gemini API key authentication via `google.golang.org/genai`
- Redis short-session memory only
- Dedicated Elasticsearch chat retrieval index

## Required Environment
- `CHAT_GEMINI_API_KEY`
- `CHAT_GEMINI_MODEL` (default `gemini-2.5-flash-lite`)
- `CHAT_SESSION_TTL_SECONDS`
- `CHAT_MAX_HISTORY_TURNS`
- `CHAT_GEMINI_TIMEOUT_SECONDS`
- `CHAT_RETRIEVAL_TIMEOUT_MS`
- `CHAT_MAX_RETRIEVAL_DOCS`
- `ELASTIC_INDEX_CHAT_RETRIEVAL`

## Local Workflow
1. Start local infra and backend runtime.
2. Build or run backend:
   - `cd backend && go test ./... -count=1`
   - `cd backend && go build ./...`
3. Rebuild the standard browse index if needed:
   - `cd backend && go run ./cmd/listing-indexer rebuild`
4. Rebuild the chat retrieval index:
   - `cd backend && go run ./cmd/listing-indexer rebuild-chat`

## Chat Endpoint
- `POST /api/chat/messages`

### Example Request
```json
{
  "session_id": "ses-rag-1",
  "message": "Ada rumah aktif di Jakarta Selatan di bawah 5 miliar?",
  "filters": {
    "transaction_type": "sale",
    "location_city": "Jakarta Selatan"
  },
  "max_documents": 5
}
```

### Example Curl
```bash
curl -s -X POST http://localhost:8080/api/chat/messages \
  -H 'Content-Type: application/json' \
  -d '{
    "session_id":"ses-rag-1",
    "message":"Ada rumah aktif di Jakarta Selatan di bawah 5 miliar?",
    "filters":{"transaction_type":"sale","location_city":"Jakarta Selatan"},
    "max_documents":5
  }'
```

## Grounding Rules
- Only active/public-safe listings may be returned
- Recommendation responses must cite concrete listing IDs/slugs
- No seller-private fields may enter prompts or API responses

## Degraded Modes
- `retrieval_unavailable`
- `generation_unavailable`
- `memory_unavailable`
- `empty_query_embedding`
- `no_active_grounded_results`

## Verification
- `cd backend && CHAT_GEMINI_API_KEY=test go test ./... -count=1`
- `cd backend && go build ./...`
