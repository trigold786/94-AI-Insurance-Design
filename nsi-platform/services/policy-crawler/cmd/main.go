package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/admin"
	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/crawler"
	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/embeddings"
	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/handler"
	authmw "github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/middleware"
	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/scheduler"
	"github.com/trigold786/94-AI-Insurance-Design/shared/config"
	"github.com/trigold786/94-AI-Insurance-Design/shared/db"
	"github.com/trigold786/94-AI-Insurance-Design/shared/middleware"
)

func main() {
	adminUser := os.Getenv("ADMIN_USERNAME")
	adminPass := os.Getenv("ADMIN_PASSWORD")
	if adminUser == "" {
		adminUser = "admin"
		log.Println("[auth] WARNING: ADMIN_USERNAME not set, using default")
	}
	if adminPass == "" {
		adminPass = "changeme"
		log.Println("[auth] WARNING: ADMIN_PASSWORD not set, using default")
	}
	adminAuth := authmw.BasicAuth(adminUser, adminPass)

	cfg := config.Load()

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	// 初始化存储层
	store, err := crawler.NewDBStore(database)
	if err != nil {
		log.Fatalf("failed to create store: %v", err)
	}

	var embedProv embeddings.EmbeddingProvider
	llmCfg, llmErr := store.GetLLMConfig()
	if llmErr == nil {
		// 优先使用独立的 Embedding 配置（火山方舟 Doubao），否则回退到 LLM 配置
		embedAPIKey := llmCfg.EmbeddingAPIKey
		embedBaseURL := llmCfg.EmbeddingEndpoint
		embedModel := llmCfg.EmbeddingModel
		embedDims := llmCfg.EmbeddingDimensions

		if embedAPIKey == "" {
			embedAPIKey = llmCfg.APIKey
		}
		if embedBaseURL == "" {
			embedBaseURL = "https://api.openai.com/v1/embeddings"
			if llmCfg.Endpoint != "" {
				if idx := strings.Index(llmCfg.Endpoint, "/chat/"); idx > 0 {
					embedBaseURL = llmCfg.Endpoint[:idx] + "/embeddings"
				}
			}
		}
		if embedModel == "" {
			embedModel = "text-embedding-3-small"
		}
		if embedDims <= 0 {
			embedDims = 1536
		}
		if embedAPIKey != "" {
			embedProv = embeddings.NewProviderFromConfig(embedAPIKey, embedBaseURL, embedModel, embedDims)
			log.Printf("[embeddings] using %s provider (dims=%d) via %s", embedProv.ModelName(), embedDims, embedBaseURL)
		}
	}
	if embedProv == nil {
		embedProv = embeddings.NewProviderFromConfig("", "", "", 1536)
		log.Println("[embeddings] no embedding API key, using hash-bow fallback")
	}

	searcher := embeddings.NewVectorSearcher(database, embedProv)

	// 初始化爬取管理器
	manager := crawler.NewCrawlerManager(store, store)

	admin.RSSFeedParser = func(data []byte) ([]admin.FeedPreviewItem, error) {
		items, err := crawler.ParseFeed(data)
		if err != nil {
			return nil, err
		}
		result := make([]admin.FeedPreviewItem, len(items))
		for i, it := range items {
			result[i] = admin.FeedPreviewItem{Title: it.Title, Link: it.Link}
		}
		return result, nil
	}

	// 可选：启用 Chromium 渲染（SPA 政府网站需要 JS 渲染）
	if os.Getenv("CHROME_ENABLED") == "true" {
		log.Println("[crawler] initializing Chrome renderer for SPA sites...")
		renderer := crawler.NewChromeRenderer()
		manager.SetRenderer(renderer)
		log.Println("[crawler] Chrome renderer enabled")
	}

	// 从 DB 加载启用的数据源配置
	sources, err := store.ListEnabledSources()
	if err != nil {
		log.Printf("[crawler] warning: could not load sources: %v", err)
	}

	watchDir := os.Getenv("POLICY_WATCH_DIR")
	if watchDir == "" {
		watchDir = "/data/policies"
	}
	manager.Init(sources, watchDir)

	// 接入差异化 scheduler，每个信源独立间隔
	crawlScheduler := scheduler.NewScheduler()
	crawlScheduler.SetTask(func(sourceID string) {
		log.Printf("[scheduler] crawl cycle for %s", sourceID)
		manager.CrawlSource(sourceID)
	})
	for _, src := range sources {
		crawlScheduler.AddSource(src.SourceID, scheduler.SourceLevel(src.SourceLevel), time.Duration(src.IntervalSec)*time.Second)
	}
	crawlScheduler.Start()

	// 立即执行一次初始全量爬取
	go func() {
		time.Sleep(3 * time.Second)
		log.Println("initial crawl cycle starting...")
		manager.CrawlAll()
	}()

	// HTTP 路由
	mux := http.NewServeMux()
	mux.Handle("/admin", adminAuth(admin.AdminPageHandler()))
	mux.Handle("/admin/dashboard", adminAuth(admin.DashboardHandler(store)))
	mux.Handle("/admin/sources", adminAuth(admin.SourceListHandler(store)))
	mux.Handle("/admin/sources/update", adminAuth(admin.SourceUpdateHandler(store)))
	mux.Handle("/admin/logs", adminAuth(admin.CrawlLogsHandler(store)))
	mux.Handle("/admin/extract-logs", adminAuth(admin.ExtractLogsHandler(store)))
	mux.Handle("/admin/regions", adminAuth(admin.RegionsHandler()))
	mux.Handle("/admin/claims", adminAuth(admin.ListClaimsHandler(store)))
	mux.Handle("/admin/claims/batch", adminAuth(admin.BatchUpdateHandler(store)))
	mux.Handle("/admin/claims/", adminAuth(admin.UpdateClaimHandler(store)))
	mux.Handle("/admin/ingest", adminAuth(admin.IngestPolicyHandler(store)))
	mux.Handle("/admin/sources/import", adminAuth(middleware.RecoveryMiddleware()(admin.SourceImportHandler(store))))
	mux.Handle("/admin/sources/create", adminAuth(admin.SourceCreateHandler(store)))
	mux.Handle("/admin/sources/delete", adminAuth(admin.SourceDeleteHandler(store)))
	mux.Handle("/admin/sources/crawl", adminAuth(admin.SourceCrawlTriggerHandler(manager)))
	mux.Handle("/admin/sources/test-rss", adminAuth(admin.RSSTestHandler()))
	mux.Handle("/admin/llm/config", adminAuth(admin.LLMConfigGetHandler(store)))
	mux.Handle("/admin/llm/config/save", adminAuth(admin.LLMConfigSaveHandler(store)))
	mux.Handle("/admin/llm/status", adminAuth(admin.LLMStatusHandler(store)))
	mux.Handle("/admin/llm/extract", adminAuth(middleware.RecoveryMiddleware()(admin.LLMExtractRunHandler(store, searcher, embedProv))))
	mux.Handle("/admin/llm/pending", adminAuth(admin.LLMPendingHandler(store)))
	mux.Handle("/admin/llm/progress", adminAuth(admin.LLMProgressHandler(store)))
	mux.Handle("/admin/pipeline", adminAuth(admin.PipelineHandler(store)))
	mux.Handle("/admin/llm/search", adminAuth(middleware.RecoveryMiddleware()(admin.AdminSearchHandler(searcher))))
	mux.Handle("/admin/search_page", adminAuth(middleware.RecoveryMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(admin.HTMLSearchPage()))
	}))))
	mux.Handle("/v1/policies/similar", middleware.RecoveryMiddleware()(handler.SimilarSearchHandler(searcher)))
	mux.Handle("/v1/policies/versions", middleware.RecoveryMiddleware()(handler.VersionsHandler(store)))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(`{"status":"ok"}`))
	})

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":39403"
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("policy-crawler starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}
