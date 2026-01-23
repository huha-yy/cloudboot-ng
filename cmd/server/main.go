package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cloudboot/cloudboot-ng/internal/api"
	"github.com/cloudboot/cloudboot-ng/internal/core/cspm"
	"github.com/cloudboot/cloudboot-ng/internal/core/logbroker"
	"github.com/cloudboot/cloudboot-ng/internal/models"
	"github.com/cloudboot/cloudboot-ng/internal/pkg/crypto"
	"github.com/cloudboot/cloudboot-ng/internal/pkg/database"
	"github.com/cloudboot/cloudboot-ng/internal/pkg/monitor"
	"github.com/cloudboot/cloudboot-ng/internal/pkg/renderer"
	"github.com/cloudboot/cloudboot-ng/web"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm/logger"
)

const (
	AppName    = "CloudBoot NG"
	AppVersion = "1.0.0-alpha"
)

func main() {
	fmt.Printf(`
╔═══════════════════════════════════════════════════════╗
║                                                       ║
║   ██████╗██╗      ██████╗ ██╗   ██╗██████╗           ║
║  ██╔════╝██║     ██╔═══██╗██║   ██║██╔══██╗          ║
║  ██║     ██║     ██║   ██║██║   ██║██║  ██║          ║
║  ██║     ██║     ██║   ██║██║   ██║██║  ██║          ║
║  ╚██████╗███████╗╚██████╔╝╚██████╔╝██████╔╝          ║
║   ╚═════╝╚══════╝ ╚═════╝  ╚═════╝ ╚═════╝           ║
║                                                       ║
║   CloudBoot NG - The Terraform for Bare Metal        ║
║   Version: %s                                  ║
║                                                       ║
╚═══════════════════════════════════════════════════════╝

`, AppVersion)

	// 初始化数据库
	dbConfig := database.Config{
		DSN:      getEnv("DB_DSN", "cloudboot.db?_journal_mode=WAL"),
		LogLevel: logger.Info,
	}

	if err := database.Init(dbConfig); err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}
	defer database.Close()

	// 初始化系统监控
	monitor.Init()
	log.Println("✅ 系统监控初始化完成")

	// 初始化LogBroker
	broker := logbroker.NewBroker()
	log.Println("✅ LogBroker初始化完成")

	// 初始化数据库备份调度器
	backupDir := getEnv("BACKUP_DIR", "./backups")
	backupInterval := getEnv("BACKUP_INTERVAL", "24h")
	interval, err := time.ParseDuration(backupInterval)
	if err != nil {
		log.Printf("⚠️  备份间隔配置无效，使用默认值24h: %v", err)
		interval = 24 * time.Hour
	}
	backupManager := database.NewBackupManager(dbConfig.DSN, backupDir)
	backupScheduler := database.NewBackupScheduler(backupManager, interval)
	backupScheduler.Start()
	log.Println("✅ 数据库备份调度器已启动")

	// 检测运行模式 (DEV=1 开发模式, 默认生产模式)
	isDev := getEnv("DEV", "") != ""

	// 初始化模板渲染器
	var templateRenderer *renderer.TemplateRenderer
	if isDev {
		// 开发模式：从文件系统加载
		log.Println("🔧 开发模式：从文件系统加载模板")
		templatesPath := "web/templates"
		templateRenderer, err = renderer.NewTemplateRenderer(templatesPath)
		if err != nil {
			log.Fatalf("❌ 模板渲染器初始化失败: %v", err)
		}
	} else {
		// 生产模式：从嵌入文件系统加载
		log.Println("📦 生产模式：从嵌入文件系统加载模板")
		templateFS, err := web.GetTemplateAssets()
		if err != nil {
			log.Fatalf("❌ 获取嵌入模板失败: %v", err)
		}
		templateRenderer, err = renderer.NewTemplateRendererFromFS(templateFS)
		if err != nil {
			log.Fatalf("❌ 模板渲染器初始化失败: %v", err)
		}
	}
	log.Println("✅ 模板渲染器初始化完成")

	// 初始化 Echo
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Renderer = templateRenderer

	// 中间件
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// 静态文件服务
	if isDev {
		// 开发模式：从文件系统提供静态文件
		log.Println("🔧 开发模式：从文件系统提供静态文件")
		e.Static("/static", "web/static")
	} else {
		// 生产模式：从嵌入文件系统提供静态文件
		log.Println("📦 生产模式：从嵌入文件系统提供静态文件")
		staticFS, err := web.GetStaticAssets()
		if err != nil {
			log.Fatalf("❌ 获取嵌入静态文件失败: %v", err)
		}
		e.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", http.FileServer(http.FS(staticFS)))))
	}

	// 路由
	setupRoutes(e, broker)

	// 启动信息
	port := getEnv("PORT", "8080")
	log.Printf("🚀 服务启动成功")
	log.Printf("📍 地址: http://localhost:%s", port)
	log.Printf("📚 API文档: http://localhost:%s/api/docs", port)
	log.Printf("🎨 Design System: http://localhost:%s/design-system", port)

	// 启动服务器
	if err := e.Start(":" + port); err != nil {
		log.Fatalf("❌ 服务启动失败: %v", err)
	}
}

func setupRoutes(e *echo.Echo, broker *logbroker.Broker) {
	// ========== DRM/安全初始化 ==========
	// TODO(生产环境): 从安全存储(HSM/Vault)加载Master Key和License
	// 当前为开发环境临时方案

	masterKey, err := crypto.GenerateAES256Key()
	if err != nil {
		log.Fatalf("❌ Master Key生成失败: %v", err)
	}
	log.Println("⚠️  开发模式: 使用临时生成的Master Key (生产环境需从HSM加载)")

	// 生成临时ECDSA密钥对 (开发环境)
	privateKey, err := crypto.GenerateECDSAKeyPair()
	if err != nil {
		log.Fatalf("❌ ECDSA密钥对生成失败: %v", err)
	}
	officialPubKey := &privateKey.PublicKey
	log.Println("⚠️  开发模式: 使用临时生成的ECDSA公钥 (生产环境需使用官方公钥)")

	// 当前License ID (开发环境使用默认值)
	currentLicenseID := getEnv("LICENSE_ID", "dev-license-00000000")
	log.Printf("📋 当前License ID: %s", currentLicenseID)

	// 初始化PluginManager (带DRM支持)
	storeDir := getEnv("STORE_DIR", "./data/store")
	pluginManager, err := cspm.NewPluginManager(storeDir, masterKey, officialPubKey, currentLicenseID)
	if err != nil {
		log.Fatalf("❌ PluginManager初始化失败: %v", err)
	}
	log.Println("✅ PluginManager初始化完成 (含DRM安全机制)")

	// 初始化Handler
	machineHandler := api.NewMachineHandler()
	jobHandler := api.NewJobHandler()
	bootHandler := api.NewBootHandler(broker)
	agentHandler := api.NewAgentHandler() // 新增：标准Agent硬件上报协议
	pxeHandler := api.NewPXEHandler(getEnv("SERVER_URL", "http://localhost:8080")) // 新增：PXE/iPXE启动
	bootConfigHandler := api.NewBootConfigHandler(getEnv("SERVER_URL", "http://localhost:8080")) // 新增：Boot配置
	streamHandler := api.NewStreamHandler(broker)
	demoHandler := api.NewDemoHandler(broker)
	profileHandler := api.NewProfileHandler()
	storeHandler := api.NewStoreHandler(pluginManager)
	webHandler := api.NewWebHandler(pluginManager)

	// 健康检查
	e.GET("/health", func(c echo.Context) error {
		// 检查数据库连接
		if err := database.HealthCheck(); err != nil {
			return c.JSON(503, map[string]string{
				"status":  "unhealthy",
				"version": AppVersion,
				"error":   err.Error(),
			})
		}

		return c.JSON(200, map[string]string{
			"status":  "ok",
			"version": AppVersion,
		})
	})


	// Design System 页面
	e.GET("/design-system", webHandler.DesignSystemPage)

	// Frontend Pages
	e.GET("/", webHandler.HomePage)
	e.GET("/machines", webHandler.MachinesPage)
	e.GET("/jobs", webHandler.JobsPage)
	e.GET("/jobs/:job_id/logs", jobLogsPageHandler)
	e.GET("/os-designer", webHandler.OSDesignerPage)
	e.GET("/store", webHandler.StorePage)
	e.GET("/settings", webHandler.SettingsPage)

	// Boot API (Agent ↔ Core)
	bootAPI := e.Group("/api/boot/v1")
	{
		// 标准硬件上报协议 (agent_handler.go)
		bootAPI.POST("/register", agentHandler.Register)   // Agent首次注册
		bootAPI.POST("/heartbeat", agentHandler.Heartbeat) // Agent心跳（定期上报）

		// 兼容老协议 (boot_handler.go)
		bootAPI.POST("/register-legacy", bootHandler.RegisterAgent)

		// 任务管理
		bootAPI.GET("/task", bootHandler.GetTask)
		bootAPI.POST("/logs", bootHandler.UploadLogs)
		bootAPI.POST("/status", bootHandler.ReportStatus)
	}

	// PXE/iPXE Boot (裸机网络启动)
	bootGroup := e.Group("/boot")
	{
		// iPXE启动脚本（HTTP Boot）
		bootGroup.GET("/ipxe/:mac", pxeHandler.ServeiPXEScript)

		// OS安装配置文件
		bootGroup.GET("/kickstart/:machine_id", bootConfigHandler.ServeKickstart)   // RHEL/CentOS
		bootGroup.GET("/autoyast/:machine_id", bootConfigHandler.ServeAutoYaST)     // SUSE/openSUSE
		// TODO: Ubuntu Autoinstall, Debian Preseed
	}

	// External API
	apiV1 := e.Group("/api/v1")
	{
		// Machine endpoints
		apiV1.GET("/machines", machineHandler.ListMachines)
		apiV1.GET("/machines/:id", machineHandler.GetMachine)
		apiV1.POST("/machines", machineHandler.CreateMachine)
		apiV1.PUT("/machines/:id", machineHandler.UpdateMachine)
		apiV1.DELETE("/machines/:id", machineHandler.DeleteMachine)
		apiV1.POST("/machines/:id/provision", machineHandler.ProvisionMachine)

		// Job endpoints
		apiV1.GET("/jobs", jobHandler.ListJobs)
		apiV1.GET("/jobs/:id", jobHandler.GetJob)
		apiV1.DELETE("/jobs/:id", jobHandler.CancelJob)

		// Profile endpoints
		apiV1.GET("/profiles", profileHandler.ListProfiles)
		apiV1.GET("/profiles/:id", profileHandler.GetProfile)
		apiV1.POST("/profiles", profileHandler.CreateProfile)
		apiV1.PUT("/profiles/:id", profileHandler.UpdateProfile)
		apiV1.DELETE("/profiles/:id", profileHandler.DeleteProfile)
		apiV1.POST("/profiles/:id/preview", profileHandler.PreviewConfig)
		apiV1.POST("/profiles/preview", profileHandler.PreviewFromPayload)

		// Store endpoints (Private Store for Provider packages)
		apiV1.POST("/store/import", storeHandler.ImportProvider)
		apiV1.GET("/store/providers", storeHandler.ListProviders)
		apiV1.GET("/store/providers/:id", storeHandler.GetProvider)
		apiV1.DELETE("/store/providers/:id", storeHandler.DeleteProvider)
	}

	// Stream API (SSE)
	e.GET("/api/stream/logs/:job_id", streamHandler.StreamLogs)

	// Demo API (演示Orchestrator执行)
	e.POST("/api/demo/orchestrator", demoHandler.TriggerOrchestratorDemo)
}

func designSystemHandler(c echo.Context) error {
	return c.HTML(200, `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Design System - CloudBoot NG</title>
    <link href="/static/css/output.css" rel="stylesheet">
    <script src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js" defer></script>
</head>
<body class="p-8">
    <div class="max-w-7xl mx-auto">
        <h1 class="text-4xl font-bold text-white mb-2">CloudBoot NG Design System</h1>
        <p class="text-slate-400 mb-8">Dark Industrial Theme - 组件库与样式指南</p>

        <!-- Colors -->
        <section class="mb-12">
            <h2 class="text-2xl font-semibold text-white mb-4">Colors</h2>
            <div class="grid grid-cols-4 gap-4">
                <div class="glass-card p-4">
                    <div class="h-16 bg-canvas rounded mb-2"></div>
                    <p class="text-sm font-mono">bg-canvas</p>
                    <p class="text-xs text-slate-500">#020617</p>
                </div>
                <div class="glass-card p-4">
                    <div class="h-16 bg-surface rounded mb-2"></div>
                    <p class="text-sm font-mono">bg-surface</p>
                    <p class="text-xs text-slate-500">#0f172a</p>
                </div>
                <div class="glass-card p-4">
                    <div class="h-16 bg-emerald-500 rounded mb-2"></div>
                    <p class="text-sm font-mono">emerald-500</p>
                    <p class="text-xs text-slate-500">Primary</p>
                </div>
                <div class="glass-card p-4">
                    <div class="h-16 bg-rose-500 rounded mb-2"></div>
                    <p class="text-sm font-mono">rose-500</p>
                    <p class="text-xs text-slate-500">Destructive</p>
                </div>
            </div>
        </section>

        <!-- Buttons -->
        <section class="mb-12">
            <h2 class="text-2xl font-semibold text-white mb-4">Buttons</h2>
            <div class="glass-card p-6 space-x-4">
                <button class="btn-primary">Primary Button</button>
                <button class="btn-destructive">Destructive Button</button>
                <button class="btn-ghost">Ghost Button</button>
            </div>
        </section>

        <!-- Badges -->
        <section class="mb-12">
            <h2 class="text-2xl font-semibold text-white mb-4">Badges</h2>
            <div class="glass-card p-6 space-x-4">
                <span class="badge badge-online">
                    <span class="dot-pulse mr-2"></span>
                    Online
                </span>
                <span class="badge badge-error">Error</span>
                <span class="badge badge-warning">Warning</span>
                <span class="badge badge-info">Info</span>
            </div>
        </section>

        <!-- Terminal -->
        <section class="mb-12">
            <h2 class="text-2xl font-semibold text-white mb-4">Matrix Terminal</h2>
            <div class="terminal">
                <div class="terminal-header">
                    <div class="flex space-x-2">
                        <div class="w-3 h-3 rounded-full bg-rose-500/20"></div>
                        <div class="w-3 h-3 rounded-full bg-amber-500/20"></div>
                        <div class="w-3 h-3 rounded-full bg-emerald-500"></div>
                    </div>
                    <div class="ml-4 text-xs text-slate-500">root@cloudboot-core: ~</div>
                </div>
                <div class="terminal-body">
                    <div class="text-emerald-500/90">> Initializing hardware probe...</div>
                    <div class="text-slate-300">> Found RAID Controller: LSI 3108</div>
                    <div class="text-emerald-500">> [RAID] Config Success ✓</div>
                    <div class="text-slate-500">> Waiting for next command...</div>
                </div>
            </div>
        </section>

        <!-- Form Inputs -->
        <section class="mb-12">
            <h2 class="text-2xl font-semibold text-white mb-4">Form Inputs</h2>
            <div class="glass-card p-6 max-w-md">
                <div class="mb-4">
                    <label class="block text-sm font-medium text-slate-400 mb-1">Hostname</label>
                    <input type="text" class="input" placeholder="server-001">
                </div>
                <div class="mb-4">
                    <label class="block text-sm font-medium text-slate-400 mb-1">MAC Address</label>
                    <input type="text" class="input font-mono" placeholder="aa:bb:cc:dd:ee:ff">
                </div>
            </div>
        </section>

        <!-- Cards -->
        <section class="mb-12">
            <h2 class="text-2xl font-semibold text-white mb-4">Cards</h2>
            <div class="grid grid-cols-3 gap-4">
                <div class="glass-card p-6">
                    <h3 class="text-lg font-medium text-white mb-2">Basic Card</h3>
                    <p class="text-slate-400">This is a basic glass card with backdrop blur effect.</p>
                </div>
                <div class="glass-card p-6">
                    <h3 class="text-lg font-medium text-white mb-2">Machine Status</h3>
                    <div class="space-y-2">
                        <div class="flex justify-between">
                            <span class="text-slate-400">CPU:</span>
                            <span class="font-mono text-slate-200">32 Cores</span>
                        </div>
                        <div class="flex justify-between">
                            <span class="text-slate-400">Memory:</span>
                            <span class="font-mono text-slate-200">128GB</span>
                        </div>
                    </div>
                </div>
                <div class="glass-card p-6">
                    <h3 class="text-lg font-medium text-white mb-2">Provider Info</h3>
                    <span class="badge badge-info">RAID Controller</span>
                    <p class="text-xs text-slate-500 mt-2 font-mono">LSI MegaRAID 3108</p>
                </div>
            </div>
        </section>
    </div>
</body>
</html>
	`)
}

func jobLogsPageHandler(c echo.Context) error {
	jobID := c.Param("job_id")

	// 获取Job信息
	var job models.Job
	if err := database.DB.Preload("Machine").First(&job, "id = ?", jobID).Error; err != nil {
		return c.String(404, "Job not found")
	}

	// 渲染job-logs页面
	return c.Render(200, "job-logs.html", map[string]interface{}{
		"job": job,
	})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
