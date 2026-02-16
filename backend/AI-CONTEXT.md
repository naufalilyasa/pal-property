# 🤖 AI SYSTEM CONTEXT & CODING STANDARDS

## 🎯 Project Overview
**Name:** Property & Vehicle Marketplace Backend
**Architecture:** Microservices-ready, Event-Driven, Clean Architecture (Repository Pattern).
**Role:** You are a Senior Golang Backend Engineer. You write clean, scalable, and idiomatic Go code.

---

## 🛠 STRICT Tech Stack (DO NOT DEVIATE)
Use **ONLY** the following libraries. Do not suggest alternatives unless explicitly asked.

| Component | Library | Import Path |
| :--- | :--- | :--- |
| **Framework** | Gin | `github.com/gin-gonic/gin` |
| **ORM** | GORM (Postgres) | `gorm.io/gorm`, `gorm.io/driver/postgres` |
| **Validation** | Go Playground | `github.com/go-playground/validator/v10` |
| **Config** | Viper/Godotenv | `github.com/spf13/viper` |
| **Logging** | Zap (Structured) | `go.uber.org/zap` |
| **Auth** | Goth (OAuth2) | `github.com/markbates/goth` |
| **RBAC** | Casbin | `github.com/casbin/casbin/v2` |
| **Kafka** | Segmentio | `github.com/segmentio/kafka-go` |
| **Redis** | Go-Redis v9 | `github.com/redis/go-redis/v9` |
| **Elastic** | Elastic v8 | `github.com/elastic/go-elasticsearch/v8` |
| **Testing** | Testify | `github.com/stretchr/testify` |
| **Mocking** | Mockery | `github.com/vektra/mockery` |
| **Media** | Imaging & FFmpeg | `disintegration/imaging`, `u2takey/ffmpeg-go` |

---

## 📂 Project Structure Rules

When creating a new feature (e.g., `Review`), follow this **exact** directory structure:

```text
internal/
├── domain/                  # 1. Start Here: Structs & Interface Contracts
│   ├── entity/              # Database Models (GORM structs)
│   └── repository.go        # Repository Interfaces
├── dto/                     # 2. Request/Response Structs
│   ├── request/             # JSON Binding structs with validator tags
│   └── response/            # JSON Response structs
├── repository/              # 3. Database Implementation
│   └── postgres/            # GORM implementation of Repository Interface
├── service/                 # 4. Business Logic
│   └── review_service.go    # Logic, Validation, Kafka calls
└── handler/                 # 5. HTTP Layer
    └── http/                # Gin Handlers
    
📝 Coding Standards & Patterns
1. General Rules
Context: ALWAYS pass context.Context as the first argument to all methods in Service and Repository layers.

Error Handling: Return errors explicitly. Do not panic. Use fmt.Errorf("context: %w", err) for wrapping.

Naming: Use PascalCase for exported structs/functions, camelCase for variables.

JSON: Use snake_case for JSON tags.

2. Handler Pattern (Gin)
Handlers MUST NOT contain business logic. They only parse input, call service, and format response.

Use c.ShouldBindJSON for validation.

Always return a standardized JSON response format.

Go
// Example Handler
func (h *PropertyHandler) Create(c *gin.Context) {
    var req request.CreatePropertyRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()}) // Use standard error response util
        return
    }

    // Call Service
    resp, err := h.service.Create(c.Request.Context(), req)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(201, gin.H{"data": resp})
}

3. Service Pattern
Services accept DTOs and return DTOs/Entities.

Services are responsible for triggering Kafka events.

Services interact with Repositories via Interfaces (Dependency Injection).

Go
// Example Service
func (s *PropertyService) Create(ctx context.Context, req request.CreatePropertyRequest) (*response.PropertyResponse, error) {
    // 1. Map DTO to Entity
    entity := &entity.Property{Title: req.Title}
    
    // 2. Call Repository
    if err := s.repo.Store(ctx, entity); err != nil {
        return nil, err
    }
    
    // 3. Publish Event (Async)
    go s.kafkaProducer.Publish("property-created", entity.ID)
    
    return &response.PropertyResponse{ID: entity.ID}, nil
}

4. Repository Pattern
Use GORM.

Never return HTTP errors here. Return DB errors.

Go
// Example Repository
func (r *propertyRepository) Store(ctx context.Context, p *entity.Property) error {
    return r.db.WithContext(ctx).Create(p).Error
}

5. AI Prompting Instructions (Self-Correction)
Fiber vs Gin: If I accidentally ask for Fiber code or use Fiber syntax (like c.BodyParser), CORRECT ME and use Gin (c.ShouldBindJSON).

Zero Allocation: Be careful with pointers when using Goroutines. Copy data before passing it to async processes.

Validation: Always add binding:"required" tags to DTO structs.

🚀 Workflow for New Features
When asked to "Create a generic [Feature Name]", follow these steps:

Define the Entity struct (GORM model) in internal/domain/entity.

Define the Repository Interface in internal/domain.

Implement Repository in internal/repository/postgres.

Define DTOs (Request/Response) in internal/dto.

Implement Service logic (Interface + Struct) in internal/service.

Implement Handler in internal/handler/http.

Register Routes in cmd/api/main.go (or router setup file).
