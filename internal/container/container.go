package container

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"zpwoot/internal/adapters/cache"
	"zpwoot/internal/adapters/database"
	"zpwoot/internal/adapters/database/repository"
	chatwootIntegration "zpwoot/internal/adapters/integration/chatwoot"
	"zpwoot/internal/adapters/integration/webhook"
	"zpwoot/internal/adapters/logger"
	"zpwoot/internal/adapters/waclient"
	"zpwoot/internal/config"
	chatwootUseCase "zpwoot/internal/core/application/usecase/chatwoot"
	"zpwoot/internal/core/application/usecase/message"
	proxyUseCase "zpwoot/internal/core/application/usecase/proxy"
	"zpwoot/internal/core/application/usecase/session"
	webhookUseCase "zpwoot/internal/core/application/usecase/webhook"
	domainChatwoot "zpwoot/internal/core/domain/chatwoot"
	domainProxy "zpwoot/internal/core/domain/proxy"
	domainSession "zpwoot/internal/core/domain/session"
	domainWebhook "zpwoot/internal/core/domain/webhook"
	"zpwoot/internal/core/ports/input"
	"zpwoot/internal/core/ports/output"
)

type Container struct {
	config   *config.Config
	logger   *logger.Logger
	database *database.Database
	migrator *database.Migrator

	cache        output.CachePort
	cacheManager *cache.CacheManager
	keyBuilder   output.CacheKeyBuilder
	serializer   output.CacheSerializer

	cachedSessionRepo  domainSession.Repository
	cachedWebhookRepo  domainWebhook.Repository
	cachedProxyRepo    domainProxy.Repository
	cachedChatwootRepo domainChatwoot.Repository

	sessionService  *domainSession.Service
	webhookService  *domainWebhook.Service
	chatwootService *domainChatwoot.Service
	proxyService    *domainProxy.Service

	whatsappClient output.WhatsAppClient
	webhookSender  output.WebhookSender

	sessionUseCases  input.SessionUseCases
	messageUseCases  input.MessageUseCases
	webhookUseCases  input.WebhookUseCases
	chatwootUseCases input.ChatwootUseCases
	proxyUseCases    input.ProxyUseCases

	chatwootIntegrator *chatwootIntegration.Integrator
}

func NewContainer(cfg *config.Config) *Container {
	return &Container{
		config: cfg,
	}
}

func (c *Container) Init() error {
	return c.InitWithContext(context.Background())
}

func (c *Container) InitWithContext(ctx context.Context) error {
	logger.Init(c.config.LogLevel)
	c.logger = logger.NewFromAppConfig(c.config)
	c.logger.Info().Msg("Initializing zpwoot container")

	c.logger.Info().Msg("Connecting to database")

	db, err := database.New(c.config, c.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	c.database = db
	c.logger.Info().Msg("Database connection established")

	c.logger.Info().Msg("Running database migrations")

	c.migrator = database.NewMigrator(db, c.logger)
	if err := c.migrator.RunMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	c.logger.Info().Msg("Database migrations completed")

	c.logger.Info().Msg("Initializing cache system")
	if err := c.initCache(); err != nil {
		return fmt.Errorf("failed to initialize cache: %w", err)
	}

	c.logger.Info().Msg("Initializing domain services with cached repositories")
	if err := c.initDomainServices(); err != nil {
		return fmt.Errorf("failed to initialize domain services: %w", err)
	}

	c.logger.Info().Msg("Initializing webhook sender")
	c.initWebhookSender()

	c.logger.Info().Msg("Initializing Chatwoot integrator")
	baseURL := buildBaseURL(c.config)
	c.chatwootUseCases = chatwootUseCase.NewUseCases(c.chatwootService, c.logger, baseURL)
	c.chatwootIntegrator = chatwootIntegration.NewIntegrator(
		c.cachedChatwootRepo,
		nil, // WhatsApp client will be set later
		c.logger,
		baseURL,
	)

	c.logger.Info().Msg("Initializing WhatsApp client")
	c.initWAClient()

	c.logger.Info().Msg("Setting WhatsApp client in Chatwoot integrator")
	c.chatwootIntegrator.SetWhatsAppClient(c.whatsappClient)

	c.logger.Info().Msg("Initializing use cases")
	c.sessionUseCases = session.NewUseCases(c.sessionService, c.whatsappClient, c.logger)
	c.messageUseCases = message.NewUseCases(c.sessionService, c.whatsappClient, c.logger)
	c.webhookUseCases = c.initWebhookUseCases()
	c.proxyUseCases = proxyUseCase.NewUseCases(c.proxyService, c.logger)

	c.logger.Info().Msg("Container initialization completed successfully")

	return nil
}

func (c *Container) initCache() error {

	c.cacheManager = cache.NewCacheManager(c.config, c.logger)

	if err := c.cacheManager.ValidateConfig(); err != nil {
		return fmt.Errorf("cache configuration validation failed: %w", err)
	}

	cacheInstance, err := c.cacheManager.CreateCache()
	if err != nil {
		return fmt.Errorf("failed to create cache instance: %w", err)
	}
	c.cache = cacheInstance

	c.keyBuilder = c.cacheManager.CreateKeyBuilder()
	c.serializer = c.cacheManager.CreateSerializer()

	c.logger.Info().
		Bool("redis_enabled", c.config.Cache.RedisEnabled).
		Str("cache_type", c.config.Cache.Type).
		Str("key_prefix", c.config.Cache.KeyPrefix).
		Msg("Cache system initialized successfully")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.cache.Ping(ctx); err != nil {
		c.logger.Warn().Err(err).Msg("Cache ping failed, but continuing...")
	} else {
		c.logger.Info().Msg("Cache connection test successful")
	}

	return nil
}

func (c *Container) initDomainServices() error {

	cacheConfig := c.cacheManager.CreateCacheConfig()

	baseSessionRepo := repository.NewSessionRepository(c.database.DB)
	baseWebhookRepo := repository.NewWebhookRepository(c.database.DB)
	baseProxyRepo := repository.NewProxyRepository(c.database.DB)
	baseChatwootRepo := repository.NewChatwootRepository(c.database.DB)

	cachedSessionRepo := cache.NewCachedSessionRepository(
		baseSessionRepo,
		c.cache,
		c.keyBuilder,
		c.serializer,
		c.logger,
		cacheConfig,
	)

	cachedWebhookRepo := cache.NewCachedWebhookRepository(
		baseWebhookRepo,
		c.cache,
		c.keyBuilder,
		c.serializer,
		c.logger,
		cacheConfig,
	)

	cachedProxyRepo := cache.NewCachedProxyRepository(
		baseProxyRepo,
		c.cache,
		c.keyBuilder,
		c.serializer,
		c.logger,
		cacheConfig,
	)

	cachedChatwootRepo := cache.NewCachedChatwootRepository(
		baseChatwootRepo,
		c.cache,
		c.keyBuilder,
		c.serializer,
		c.logger,
		cacheConfig,
	)

	c.sessionService = domainSession.NewService(cachedSessionRepo)
	c.webhookService = domainWebhook.NewService()
	c.proxyService = domainProxy.NewService(cachedProxyRepo)
	c.chatwootService = domainChatwoot.NewService(cachedChatwootRepo)

	c.cachedSessionRepo = cachedSessionRepo
	c.cachedWebhookRepo = cachedWebhookRepo
	c.cachedProxyRepo = cachedProxyRepo
	c.cachedChatwootRepo = cachedChatwootRepo

	c.logger.Info().Msg("Domain services initialized with cached repositories")
	return nil
}

func (c *Container) initWAClient() {

	baseSessionRepo := repository.NewSessionRepository(c.database.DB)
	sessionRepo := repository.NewSessionRepo(baseSessionRepo)

	waContainer := waclient.NewWAStoreContainer(
		c.database.DB,
		c.logger,
		c.config.Database.URL,
	)
	waClient := waclient.NewWAClient(waContainer, c.logger, sessionRepo, c.webhookSender, c.cachedWebhookRepo, c.chatwootIntegrator)
	c.whatsappClient = waclient.NewWAClientAdapter(waClient)
}

func (c *Container) Start(ctx context.Context) error {
	return c.InitWithContext(ctx)
}

func (c *Container) Stop(ctx context.Context) error {
	c.logger.Info().Msg("Stopping zpwoot container")

	if c.cache != nil {
		if err := c.cache.Close(); err != nil {
			c.logger.Error().Err(err).Msg("Failed to close cache connection")
		} else {
			c.logger.Info().Msg("Cache connection closed")
		}
	}

	if c.database != nil {
		if err := c.database.Close(); err != nil {
			c.logger.Error().Err(err).Msg("Failed to close database connection")
			return err
		}
		c.logger.Info().Msg("Database connection closed")
	}

	c.logger.Info().Msg("Container stopped successfully")
	return nil
}

func (c *Container) GetConfig() *config.Config {
	return c.config
}

func (c *Container) GetLogger() *logger.Logger {
	return c.logger
}

func (c *Container) GetDatabase() *database.Database {
	return c.database
}

func (c *Container) GetMigrator() *database.Migrator {
	return c.migrator
}

func (c *Container) GetSessionService() *domainSession.Service {
	return c.sessionService
}

func (c *Container) GetWhatsAppClient() output.WhatsAppClient {
	return c.whatsappClient
}

func (c *Container) GetSessionUseCases() input.SessionUseCases {
	return c.sessionUseCases
}

func (c *Container) GetMessageUseCases() input.MessageUseCases {
	return c.messageUseCases
}

func (c *Container) GetWebhookUseCases() input.WebhookUseCases {
	return c.webhookUseCases
}

func (c *Container) GetProxyUseCases() input.ProxyUseCases {
	return c.proxyUseCases
}

func (c *Container) GetChatwootUseCases() input.ChatwootUseCases {
	return c.chatwootUseCases
}

func (c *Container) GetChatwootWebhookHandler() input.ChatwootWebhookHandler {
	return chatwootIntegration.NewWebhookAdapter(c.chatwootIntegrator.GetWebhookHandler(), c.logger)
}

func (c *Container) GetChatwootIntegrator() *chatwootIntegration.Integrator {
	return c.chatwootIntegrator
}

func buildBaseURL(cfg *config.Config) string {

	if cfg.ServerURL != "" {
		return cfg.ServerURL
	}

	if cfg.GlobalWebhookURL != "" {
		return cfg.GlobalWebhookURL
	}

	if cfg.IsDevelopment() {
		return "http://localhost:" + cfg.Port
	}

	host := cfg.ServerHost
	if host == "0.0.0.0" {
		host = "localhost"
	}

	return "http://" + host + ":" + cfg.Port
}

func (c *Container) GetCache() output.CachePort {
	return c.cache
}

func (c *Container) GetCacheManager() *cache.CacheManager {
	return c.cacheManager
}

func (c *Container) GetKeyBuilder() output.CacheKeyBuilder {
	return c.keyBuilder
}

func (c *Container) GetSerializer() output.CacheSerializer {
	return c.serializer
}

func (c *Container) GetCachedSessionRepo() domainSession.Repository {
	return c.cachedSessionRepo
}

func (c *Container) GetCachedWebhookRepo() domainWebhook.Repository {
	return c.cachedWebhookRepo
}

func (c *Container) GetCachedProxyRepo() domainProxy.Repository {
	return c.cachedProxyRepo
}

func (c *Container) GetCachedChatwootRepo() domainChatwoot.Repository {
	return c.cachedChatwootRepo
}

func (c *Container) GetWebhookSender() output.WebhookSender {
	return c.webhookSender
}

func (c *Container) initWebhookSender() {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	c.webhookSender = webhook.NewHTTPWebhookSender(httpClient, c.logger)
}

func (c *Container) initWebhookUseCases() input.WebhookUseCases {
	webhookRepo := repository.NewWebhookRepository(c.database.DB)
	return webhookUseCase.NewWebhookUseCases(webhookRepo, c.webhookService)
}
