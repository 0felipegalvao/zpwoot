# 🎯 Análise Completa: Estratégia de Cache com Redis

## 📊 Mapeamento de Todas as Consultas ao Banco

### 🔴 CRÍTICO - Alta Frequência (> 100/min)

#### 1. **Session.GetByID** - Em Eventos WhatsApp
```go
// events.go:319 - Chamado em CADA evento de mensagem
sess, err := eh.sessionRepo.GetByID(ctx, sessionID)
```

**Contexto:**
- Chamado quando recebe mensagem do WhatsApp
- Precisa da sessão para processar evento
- **Frequência:** 100-1000 eventos/min por sessão ativa

**Dados consultados:**
- `id`, `name`, `apiKey`, `deviceJid`, `isConnected`
- `connectionError`, `qrCode`, `qrCodeExpiresAt`
- `createdAt`, `updatedAt`

**Cache?** ✅ **SIM - CRÍTICO**

---

#### 2. **Webhook.GetBySessionID** - Em Eventos WhatsApp
```go
// events.go:218 - Chamado em CADA evento para enviar webhook
webhookConfig, err := eh.webhookRepo.GetBySessionID(ctx, client.SessionID)
```

**Contexto:**
- Chamado para verificar se deve enviar webhook
- **Frequência:** 100-1000 eventos/min por sessão ativa

**Dados consultados:**
- `id`, `sessionId`, `url`, `secret`, `events`, `enabled`

**Cache?** ✅ **SIM - CRÍTICO**

---

### 🟡 MÉDIO - Frequência Moderada (10-100/min)

#### 3. **Session.GetByID** - Em Use Cases HTTP
```go
// Chamado em:
// - manager.go:169 (LoadSession)
// - manager.go:224 (ConnectSession)
// - manager.go:510 (handlePairSuccess)
// - Todos os use cases de session
```

**Contexto:**
- Chamado em requisições HTTP
- **Frequência:** 10-50 req/min

**Cache?** ✅ **SIM - MÉDIO**

---

#### 4. **Webhook.GetBySessionID** - Em Use Cases HTTP
```go
// Chamado em:
// - webhook/get.go:21
// - webhook/create.go:52
// - webhook/update.go:44
// - webhook/upsert.go:44
```

**Contexto:**
- Operações CRUD de webhook
- **Frequência:** 1-10 req/min

**Cache?** ⚠️ **OPCIONAL** (já coberto pelo cache de eventos)

---

### 🟢 BAIXO - Frequência Baixa (< 10/min)

#### 5. **Session.GetByName**
```go
// session/service.go:76 - Validação de nome único
existing, err := s.repo.GetByName(ctx, name)
```

**Contexto:**
- Apenas ao criar sessão
- **Frequência:** < 5/dia

**Cache?** ❌ **NÃO**

---

#### 6. **Session.GetByJID**
```go
// Não usado atualmente no código
```

**Cache?** ❌ **NÃO**

---

## 🎯 Estratégia de Cache Proposta

### Arquitetura: Redis com 2 Caches Separados

```
┌─────────────────────────────────────────────────────┐
│                  Application                        │
│  ┌──────────────────────────────────────────────┐  │
│  │         Event Handler / Use Cases            │  │
│  └──────────────────────────────────────────────┘  │
│                       │                             │
│                       ▼                             │
│  ┌──────────────────────────────────────────────┐  │
│  │           Cache Layer (Redis)                │  │
│  │  ┌────────────────┐  ┌──────────────────┐   │  │
│  │  │ Session Cache  │  │ Webhook Cache    │   │  │
│  │  │ TTL: 30s       │  │ TTL: 5min        │   │  │
│  │  └────────────────┘  └──────────────────┘   │  │
│  └──────────────────────────────────────────────┘  │
│                       │                             │
│                       ▼                             │
│  ┌──────────────────────────────────────────────┐  │
│  │         PostgreSQL Database                  │  │
│  └──────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

---

## 📋 Cache 1: Session Cache

### Configuração
```yaml
cache:
  session:
    enabled: true
    ttl: 30s              # TTL curto (dados mudam frequentemente)
    key_prefix: "session:"
    invalidate_on_update: true
```

### Chave Redis
```
session:{sessionID} → JSON da Session
```

### Exemplo
```redis
SET session:abc-123 '{"id":"abc-123","name":"my-session","isConnected":true,...}' EX 30
```

### Quando Cachear?
- ✅ `GetByID` - Sempre
- ❌ `GetByName` - Não (baixa frequência)
- ❌ `GetByJID` - Não (não usado)

### Quando Invalidar?
```go
// Invalidar quando:
- session.Update()
- session.SetConnected()
- session.SetDisconnected()
- session.SetError()
- session.SetQRCode()
```

### TTL: 30 segundos
**Por quê?**
- Dados mudam frequentemente (status, QR, connected)
- TTL curto garante dados relativamente atualizados
- Mesmo com TTL curto, reduz 90% das queries

---

## 📋 Cache 2: Webhook Cache

### Configuração
```yaml
cache:
  webhook:
    enabled: true
    ttl: 5m               # TTL longo (dados raramente mudam)
    key_prefix: "webhook:"
    invalidate_on_update: true
```

### Chave Redis
```
webhook:{sessionID} → JSON do Webhook
```

### Exemplo
```redis
SET webhook:abc-123 '{"id":"xyz","sessionId":"abc-123","url":"https://...","enabled":true}' EX 300
```

### Quando Cachear?
- ✅ `GetBySessionID` - Sempre

### Quando Invalidar?
```go
// Invalidar quando:
- webhook.Create()
- webhook.Update()
- webhook.Delete()
- webhook.Upsert()
```

### TTL: 5 minutos
**Por quê?**
- Dados raramente mudam
- TTL longo maximiza hit rate
- Invalidação explícita garante consistência

---

## 🏗️ Implementação

### 1. Interface Genérica de Cache

```go
// internal/core/ports/output/cache.go
package output

type Cache interface {
    Get(ctx context.Context, key string, dest interface{}) error
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Clear(ctx context.Context, pattern string) error
    GetMetrics() CacheMetrics
}
```

### 2. Adapter Redis

```go
// internal/adapters/cache/redis_cache.go
package cache

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/redis/go-redis/v9"
)

type RedisCache struct {
    client     *redis.Client
    keyPrefix  string
    defaultTTL time.Duration
}

func NewRedisCache(client *redis.Client, keyPrefix string, ttl time.Duration) *RedisCache {
    return &RedisCache{
        client:     client,
        keyPrefix:  keyPrefix,
        defaultTTL: ttl,
    }
}

func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
    fullKey := c.keyPrefix + key
    
    data, err := c.client.Get(ctx, fullKey).Bytes()
    if err == redis.Nil {
        return nil // Cache miss
    }
    if err != nil {
        return err
    }
    
    return json.Unmarshal(data, dest)
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    fullKey := c.keyPrefix + key
    
    if ttl == 0 {
        ttl = c.defaultTTL
    }
    
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    
    return c.client.Set(ctx, fullKey, data, ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
    fullKey := c.keyPrefix + key
    return c.client.Del(ctx, fullKey).Err()
}
```

### 3. Session Cache Wrapper

```go
// internal/adapters/cache/session_cache.go
package cache

import (
    "context"
    "zpwoot/internal/core/domain/session"
    "zpwoot/internal/core/ports/output"
)

type SessionCache struct {
    cache output.Cache
    repo  session.Repository
}

func NewSessionCache(cache output.Cache, repo session.Repository) *SessionCache {
    return &SessionCache{
        cache: cache,
        repo:  repo,
    }
}

func (c *SessionCache) GetByID(ctx context.Context, id string) (*session.Session, error) {
    // 1. Try cache
    var sess session.Session
    err := c.cache.Get(ctx, id, &sess)
    if err == nil && sess.ID != "" {
        return &sess, nil
    }
    
    // 2. Cache miss - get from DB
    sess, err := c.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // 3. Store in cache
    _ = c.cache.Set(ctx, id, sess, 30*time.Second)
    
    return sess, nil
}

func (c *SessionCache) Invalidate(ctx context.Context, id string) error {
    return c.cache.Delete(ctx, id)
}
```

### 4. Webhook Cache Wrapper

```go
// internal/adapters/cache/webhook_cache.go
package cache

import (
    "context"
    "zpwoot/internal/core/domain/webhook"
    "zpwoot/internal/core/ports/output"
)

type WebhookCache struct {
    cache output.Cache
    repo  webhook.Repository
}

func NewWebhookCache(cache output.Cache, repo webhook.Repository) *WebhookCache {
    return &WebhookCache{
        cache: cache,
        repo:  repo,
    }
}

func (c *WebhookCache) GetBySessionID(ctx context.Context, sessionID string) (*webhook.Webhook, error) {
    // 1. Try cache
    var wh webhook.Webhook
    err := c.cache.Get(ctx, sessionID, &wh)
    if err == nil && wh.ID != "" {
        return &wh, nil
    }
    
    // 2. Cache miss - get from DB
    wh, err := c.repo.GetBySessionID(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    // 3. Store in cache
    _ = c.cache.Set(ctx, sessionID, wh, 5*time.Minute)
    
    return wh, nil
}

func (c *WebhookCache) Invalidate(ctx context.Context, sessionID string) error {
    return c.cache.Delete(ctx, sessionID)
}
```

---

## 📊 Impacto Estimado

### Sem Cache (Atual)

```
Session queries:
- Eventos: 1000/min × 10 sessões = 10,000 queries/min
- HTTP: 50/min × 10 sessões = 500 queries/min
Total Session: 10,500 queries/min

Webhook queries:
- Eventos: 1000/min × 10 sessões = 10,000 queries/min
- HTTP: 10/min × 10 sessões = 100 queries/min
Total Webhook: 10,100 queries/min

TOTAL: 20,600 queries/min ao PostgreSQL
```

### Com Cache Redis

```
Session queries (TTL 30s, ~90% hit rate):
- Eventos: 1000/min × 10% = 1,000 queries/min
- HTTP: 50/min × 10% = 50 queries/min
Total Session: 1,050 queries/min

Webhook queries (TTL 5min, ~99% hit rate):
- Eventos: 1000/min × 1% = 100 queries/min
- HTTP: 10/min × 1% = 1 queries/min
Total Webhook: 101 queries/min

TOTAL: 1,151 queries/min ao PostgreSQL
```

**Redução:** 95% menos queries (20,600 → 1,151)

---

## ⚙️ Configuração

### docker-compose.yml
```yaml
services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes
```

### config.yaml
```yaml
redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
  pool_size: 10

cache:
  session:
    enabled: true
    ttl: 30s
    key_prefix: "zpwoot:session:"
  
  webhook:
    enabled: true
    ttl: 5m
    key_prefix: "zpwoot:webhook:"
```

---

## 🎯 Resumo da Estratégia

| Cache | Dados | TTL | Hit Rate | Invalidação |
|-------|-------|-----|----------|-------------|
| **Session** | Session entity | 30s | ~90% | Ao atualizar session |
| **Webhook** | Webhook config | 5min | ~99% | Ao atualizar webhook |

**Benefícios:**
- ✅ 95% redução de queries ao PostgreSQL
- ✅ Latência de eventos cai de 10-50ms para 1-5ms
- ✅ Compartilhado entre múltiplas instâncias
- ✅ Persiste entre restarts
- ✅ Escalável horizontalmente

**Trade-offs:**
- ⚠️ Latência de rede Redis (1-5ms vs < 1µs memória)
- ⚠️ Dependência externa (Redis)
- ⚠️ Dados podem ficar desatualizados (TTL)

---

Quer que eu implemente essa estratégia completa com Redis? 🚀

