# Validação de Dados - zpwoot

## 📋 Visão Geral

O zpwoot implementa validação robusta de dados em todas as camadas da aplicação usando a biblioteca **go-playground/validator/v10**. A validação é aplicada de forma consistente em DTOs, Use Cases e Handlers HTTP.

## 🏗️ Arquitetura de Validação

### **Camadas de Validação**

```
HTTP Request
    ↓
Handler (Validação Básica)
    ↓
Use Case (Validação com validator)
    ↓
Domain Service (Validação de Regras de Negócio)
    ↓
Repository
```

### **Localização dos Componentes**

- **Validador Central**: `internal/core/application/validator/validator.go`
- **HTTP Helpers**: `internal/core/application/validator/http_errors.go`
- **DTOs com Tags**: `internal/core/application/dto/*.go`
- **Use Cases**: `internal/core/application/usecase/*/`
- **Handlers**: `internal/adapters/http/handlers/*.go`

---

## 🎯 Validadores Customizados

### **1. Phone Number (`phone`)**
Valida números de telefone no formato E.164.

```go
Phone string `json:"phone" validate:"required,phone"`
```

**Exemplos Válidos**:
- `5511999999999`
- `+5511999999999`
- `+1234567890`

**Exemplos Inválidos**:
- `123` (muito curto)
- `abc123` (contém letras)
- `00000000000` (não começa com dígito válido)

---

### **2. WhatsApp JID (`jid`)**
Valida identificadores WhatsApp (Java ID).

```go
To string `json:"to" validate:"required,jid"`
```

**Exemplos Válidos**:
- `5511999999999@s.whatsapp.net`
- `120363123456789012@g.us`
- `123456789@broadcast`
- `123456789@newsletter`

**Exemplos Inválidos**:
- `5511999999999` (falta sufixo)
- `invalid@domain.com` (domínio inválido)

---

### **3. WhatsApp URL (`whatsapp_url`)**
Valida URLs HTTP/HTTPS.

```go
URL string `json:"url" validate:"required,whatsapp_url"`
```

**Exemplos Válidos**:
- `https://api.example.com/webhook`
- `http://localhost:3000/webhook`

**Exemplos Inválidos**:
- `ftp://example.com` (protocolo inválido)
- `example.com` (falta protocolo)
- `not a url` (formato inválido)

---

### **4. Session ID (`session_id`)**
Valida IDs de sessão (alfanumérico, dash, underscore).

```go
Name string `json:"name" validate:"required,session_id"`
```

**Exemplos Válidos**:
- `my-session`
- `session_123`
- `MySession-2024`

**Exemplos Inválidos**:
- `my session` (contém espaço)
- `session@123` (caractere inválido)
- `session#1` (caractere especial)

---

### **5. Webhook Event (`webhook_event`)**
Valida tipos de eventos de webhook.

```go
Events []string `json:"events" validate:"omitempty,dive,webhook_event"`
```

**Eventos Válidos**:
- `Message`, `MessageRevoked`, `MessageReaction`
- `Connected`, `Disconnected`, `QRCode`
- `GroupInfo`, `JoinedGroup`
- `CallOffer`, `CallAccept`
- E mais 20+ eventos...

---

### **6. Message Type (`message_type`)**
Valida tipos de mensagem.

```go
Type string `json:"type" validate:"required,message_type"`
```

**Tipos Válidos**:
- `text`, `image`, `video`, `audio`
- `document`, `sticker`, `location`
- `contact`, `reaction`, `poll`
- `buttons`, `list`, `template`

---

### **7. Presence Type (`presence_type`)**
Valida tipos de presença.

```go
Presence string `json:"presence" validate:"required,presence_type"`
```

**Tipos Válidos**:
- `available`, `unavailable`
- `composing`, `recording`, `paused`

---

## 📝 Exemplos de Uso

### **1. DTO com Validação**

```go
type CreateRequest struct {
    Name     string           `json:"name" validate:"required,min=1,max=100,session_id"`
    APIKey   string           `json:"apiKey,omitempty" validate:"omitempty,len=32,alphanum"`
    Settings *SessionSettings `json:"settings,omitempty" validate:"omitempty,dive"`
}
```

### **2. Use Case com Validação**

```go
func (uc *CreateUseCase) Execute(ctx context.Context, req *dto.CreateRequest) (*dto.CreateSessionResponse, error) {
    // Validate request using validator
    if err := validator.Validate(req); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    // Additional custom validation
    if err := req.Validate(); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    // ... rest of the logic
}
```

### **3. Handler HTTP com Validação**

```go
func (h *SessionHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req dto.CreateRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.writeErrorResponse(w, http.StatusBadRequest, dto.ErrorCodeBadRequest, "Invalid JSON body")
        return
    }

    // Validate request using validator
    if err := validator.Validate(&req); err != nil {
        validator.WriteValidationError(w, err)
        return
    }

    // ... call use case
}
```

---

## 🔍 Respostas de Erro

### **Formato de Erro de Validação**

```json
{
  "error": "validation_error",
  "message": "Request validation failed",
  "errors": [
    {
      "field": "Phone",
      "message": "Phone must be a valid phone number (E.164 format)",
      "tag": "phone",
      "value": "abc123"
    },
    {
      "field": "Text",
      "message": "Text is required",
      "tag": "required",
      "value": ""
    }
  ]
}
```

### **Mensagens de Erro Amigáveis**

O validador converte tags técnicas em mensagens amigáveis:

| Tag | Mensagem |
|-----|----------|
| `required` | `{field} is required` |
| `min` | `{field} must be at least {param} characters` |
| `max` | `{field} must be at most {param} characters` |
| `email` | `{field} must be a valid email address` |
| `url` | `{field} must be a valid URL` |
| `phone` | `{field} must be a valid phone number (E.164 format)` |
| `jid` | `{field} must be a valid WhatsApp JID` |

---

## 🧪 Testando Validação

### **Teste 1: Criar Sessão com Nome Inválido**

```bash
curl -X POST http://localhost:8080/sessions \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my session!"
  }'
```

**Resposta Esperada**:
```json
{
  "error": "validation_error",
  "message": "Request validation failed",
  "errors": [
    {
      "field": "Name",
      "message": "Name must contain only alphanumeric characters, dashes, and underscores",
      "tag": "session_id"
    }
  ]
}
```

---

### **Teste 2: Enviar Mensagem com Telefone Inválido**

```bash
curl -X POST http://localhost:8080/sessions/my-session/messages/send/text \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "abc123",
    "text": "Hello!"
  }'
```

**Resposta Esperada**:
```json
{
  "error": "validation_error",
  "message": "Request validation failed",
  "errors": [
    {
      "field": "Phone",
      "message": "Phone must be a valid phone number (E.164 format)",
      "tag": "phone",
      "value": "abc123"
    }
  ]
}
```

---

### **Teste 3: Configurar Webhook com URL Inválida**

```bash
curl -X POST http://localhost:8080/sessions/my-session/webhooks \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "not-a-url",
    "events": ["InvalidEvent"]
  }'
```

**Resposta Esperada**:
```json
{
  "error": "validation_error",
  "message": "Request validation failed",
  "errors": [
    {
      "field": "URL",
      "message": "URL must be a valid HTTP/HTTPS URL",
      "tag": "whatsapp_url"
    },
    {
      "field": "Events[0]",
      "message": "Events[0] must be a valid webhook event type",
      "tag": "webhook_event"
    }
  ]
}
```

---

## 📚 Tags de Validação Disponíveis

### **Tags Padrão (go-playground/validator)**

- `required` - Campo obrigatório
- `omitempty` - Valida apenas se não vazio
- `min=N` - Tamanho/valor mínimo
- `max=N` - Tamanho/valor máximo
- `len=N` - Tamanho exato
- `gt=N` - Maior que
- `gte=N` - Maior ou igual
- `lt=N` - Menor que
- `lte=N` - Menor ou igual
- `oneof=a b c` - Um dos valores
- `email` - Email válido
- `url` - URL válida
- `numeric` - Apenas números
- `alphanum` - Alfanumérico
- `base64` - Base64 válido
- `dive` - Valida elementos de array/slice

### **Tags Customizadas (zpwoot)**

- `phone` - Telefone E.164
- `jid` - WhatsApp JID
- `whatsapp_url` - HTTP/HTTPS URL
- `session_id` - ID de sessão válido
- `webhook_event` - Evento de webhook válido
- `message_type` - Tipo de mensagem válido
- `presence_type` - Tipo de presença válido

---

## ✅ Benefícios

1. ✅ **Validação Consistente**: Mesmas regras em todas as camadas
2. ✅ **Mensagens Amigáveis**: Erros claros e descritivos
3. ✅ **Segurança**: Previne dados inválidos no sistema
4. ✅ **Manutenibilidade**: Fácil adicionar novas validações
5. ✅ **Documentação**: Tags servem como documentação
6. ✅ **Performance**: Validação rápida e eficiente
7. ✅ **Testabilidade**: Fácil testar validações

---

## 🔧 Adicionando Novos Validadores

### **1. Criar Função de Validação**

```go
// internal/core/application/validator/validator.go

func validateCustomField(fl validator.FieldLevel) bool {
    value := fl.Field().String()
    // Sua lógica de validação aqui
    return true // ou false
}
```

### **2. Registrar Validador**

```go
func init() {
    validate = validator.New()
    _ = validate.RegisterValidation("custom_field", validateCustomField)
}
```

### **3. Adicionar Mensagem de Erro**

```go
func getErrorMessage(e validator.FieldError) string {
    // ...
    case "custom_field":
        return fmt.Sprintf("%s must be a valid custom field", field)
    // ...
}
```

### **4. Usar em DTOs**

```go
type MyDTO struct {
    Field string `json:"field" validate:"required,custom_field"`
}
```

---

## 📖 Referências

- [go-playground/validator](https://github.com/go-playground/validator)
- [Validator Documentation](https://pkg.go.dev/github.com/go-playground/validator/v10)
- [E.164 Phone Format](https://en.wikipedia.org/wiki/E.164)

