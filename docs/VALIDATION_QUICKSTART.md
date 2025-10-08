# 🚀 Validação - Quick Start

## 📋 Guia Rápido de Uso

Este guia mostra como usar o sistema de validação do zpwoot em 5 minutos.

---

## 🎯 Validadores Disponíveis

| Validador | Descrição | Exemplo |
|-----------|-----------|---------|
| `phone` | Telefone E.164 | `5511999999999` |
| `jid` | WhatsApp JID | `5511999999999@s.whatsapp.net` |
| `whatsapp_url` | URL HTTP/HTTPS | `https://example.com/webhook` |
| `session_id` | ID de sessão | `my-session` |
| `webhook_event` | Evento webhook | `Message`, `Connected` |
| `message_type` | Tipo de mensagem | `text`, `image`, `video` |
| `presence_type` | Tipo de presença | `available`, `composing` |

---

## 💻 Uso em DTOs

### **Exemplo 1: Validação Simples**

```go
type CreateRequest struct {
    Name string `json:"name" validate:"required,min=3,max=50,session_id"`
}
```

### **Exemplo 2: Validação Condicional**

```go
type SendMessageRequest struct {
    Phone string `json:"phone" validate:"required,phone"`
    Text  string `json:"text" validate:"required_if=Type text,min=1,max=65536"`
}
```

### **Exemplo 3: Validação de Array**

```go
type WebhookRequest struct {
    Events []string `json:"events" validate:"omitempty,dive,webhook_event"`
}
```

### **Exemplo 4: Validação de Struct Aninhado**

```go
type CreateRequest struct {
    Settings *SessionSettings `json:"settings" validate:"omitempty,dive"`
}
```

---

## 🔧 Uso em Use Cases

```go
import "zpwoot/internal/core/application/validator"

func (uc *CreateUseCase) Execute(ctx context.Context, req *dto.CreateRequest) error {
    // Validar request
    if err := validator.Validate(req); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    // Processar request...
}
```

---

## 🌐 Uso em Handlers HTTP

```go
import "zpwoot/internal/core/application/validator"

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    var req dto.CreateRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        // Handle JSON decode error
        return
    }
    
    // Validar request
    if err := validator.Validate(&req); err != nil {
        validator.WriteValidationError(w, err)
        return
    }
    
    // Processar request...
}
```

---

## 📝 Tags de Validação Comuns

### **Campos Obrigatórios**
```go
Field string `validate:"required"`
```

### **Tamanho Mínimo/Máximo**
```go
Name string `validate:"min=3,max=50"`
```

### **Valores Permitidos**
```go
Type string `validate:"oneof=text image video"`
```

### **Validação Condicional**
```go
Text string `validate:"required_if=Type text"`
```

### **Validação de Array**
```go
Events []string `validate:"dive,webhook_event"`
```

### **Validação de Struct**
```go
Settings *Settings `validate:"dive"`
```

### **Campo Opcional**
```go
Caption string `validate:"omitempty,max=1024"`
```

---

## 🔍 Formato de Erro

### **Request**
```bash
POST /sessions
{
  "name": "my session!"
}
```

### **Response (400)**
```json
{
  "error": "validation_error",
  "message": "Request validation failed",
  "errors": [
    {
      "field": "Name",
      "message": "Name must contain only alphanumeric characters, dashes, and underscores",
      "tag": "session_id",
      "value": "my session!"
    }
  ]
}
```

---

## 🧪 Testando

### **Teste 1: Nome Inválido**
```bash
curl -X POST http://localhost:8080/sessions \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"name": "my session!"}'
```

**Esperado**: Erro de validação (session_id)

### **Teste 2: Telefone Inválido**
```bash
curl -X POST http://localhost:8080/sessions/test/messages/send/text \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"phone": "abc", "text": "Hello"}'
```

**Esperado**: Erro de validação (phone)

### **Teste 3: URL Inválida**
```bash
curl -X POST http://localhost:8080/sessions/test/webhooks \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"url": "not-a-url"}'
```

**Esperado**: Erro de validação (whatsapp_url)

---

## 🎨 Criando Validador Customizado

### **1. Criar Função**
```go
// internal/core/application/validator/validator.go

func validateCustom(fl validator.FieldLevel) bool {
    value := fl.Field().String()
    // Sua lógica aqui
    return true
}
```

### **2. Registrar**
```go
func init() {
    validate = validator.New()
    _ = validate.RegisterValidation("custom", validateCustom)
}
```

### **3. Adicionar Mensagem**
```go
func getErrorMessage(e validator.FieldError) string {
    switch e.Tag() {
    case "custom":
        return fmt.Sprintf("%s must be valid", e.Field())
    }
}
```

### **4. Usar**
```go
type MyDTO struct {
    Field string `validate:"required,custom"`
}
```

---

## 📚 Documentação Completa

- **Guia Completo**: `docs/VALIDATION.md`
- **Implementação**: `docs/VALIDATION_IMPLEMENTATION.md`
- **Resumo**: `VALIDATION_SUMMARY.md`
- **Testes**: `examples/validation_tests.sh`

---

## ✅ Checklist de Validação

Ao criar um novo DTO:

- [ ] Adicionar tags `validate` em todos os campos
- [ ] Usar validadores customizados quando apropriado
- [ ] Adicionar validação no handler HTTP
- [ ] Adicionar validação no use case
- [ ] Testar com dados inválidos
- [ ] Verificar mensagens de erro

---

## 🎯 Exemplos Práticos

### **Session**
```go
type CreateRequest struct {
    Name   string `validate:"required,min=1,max=100,session_id"`
    APIKey string `validate:"omitempty,len=32,alphanum"`
}
```

### **Message**
```go
type SendTextRequest struct {
    Phone string `validate:"required,phone"`
    Text  string `validate:"required,min=1,max=65536"`
}
```

### **Webhook**
```go
type CreateWebhookRequest struct {
    URL    string   `validate:"required,whatsapp_url"`
    Events []string `validate:"omitempty,dive,webhook_event"`
    Secret string   `validate:"omitempty,min=8,max=256"`
}
```

### **Location**
```go
type Location struct {
    Latitude  float64 `validate:"required,min=-90,max=90"`
    Longitude float64 `validate:"required,min=-180,max=180"`
}
```

---

## 🚀 Pronto para Usar!

O sistema de validação está **100% funcional** e pronto para uso em produção!

**Próximos Passos**:
1. Leia a documentação completa em `docs/VALIDATION.md`
2. Execute os testes em `examples/validation_tests.sh`
3. Comece a usar validação em seus DTOs!

---

**zpwoot** - Validação robusta e fácil de usar! ✅

