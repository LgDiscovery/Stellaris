package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"kubeai/internal/jobscheduler/handler"
	"kubeai/internal/jobscheduler/repository"
	"kubeai/internal/jobscheduler/service"
	"kubeai/internal/model"
	"kubeai/internal/pkg/k8s"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	db  *gorm.DB
	err error
)

func initConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath("../../configs")
	viper.SetConfigType("yaml")
	if err = viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %s", err)
	}
}

func initDB() {
	// 2. 连接数据库
	host := viper.GetString("database.host")
	port := viper.GetInt("database.port")
	user := viper.GetString("database.user")
	password := viper.GetString("database.password")
	dbname := viper.GetString("database.dbname")
	sslmode := viper.GetString("database.sslmode")
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %s", err)
	}
	log.Println("数据库连接成功")

}

// syncJobStatus 任务状态同步定时任务
func syncJobStatus(ctx context.Context, jobSvc *service.JobService) {
	// 30秒执行一次
	ticker := time.NewTicker(time.Duration(viper.GetInt("sync.jobscheduler.status.intervalTime")) * time.Second)
	defer ticker.Stop() // 函数退出时关闭定时器，避免内存泄漏
	log.Println("任务状态同步协程启动成功，每30秒同步一次非终态任务状态")

	for {
		select {
		case <-ctx.Done():
			// 收到退出信号，停止同步任务
			log.Println("任务状态同步协程正在退出...")
			return
		case <-ticker.C:
			// 同步逻辑：拉取所有 待处理/运行中 任务，同步K8s状态
			log.Println("开始同步非终态任务状态...")
			// 分页拉取所有非终态任务（page=1，size=1000 满足大部分场景）
			jobs, _, err := jobSvc.ListJobs(1, 1000, "pending,running")
			if err != nil {
				log.Printf("获取未完成任务失败: %v", err)
				continue
			}

			// 遍历同步每个任务状态
			for _, job := range jobs {
				if err := jobSvc.SyncJobStatus(ctx, job.ID); err != nil {
					log.Printf("同步任务[%s]状态失败: %v", job.ID, err)
				}
			}
			log.Printf("任务状态同步完成，共同步 %d 个任务\n", len(jobs))
		}
	}

}

func main() {
	// 1. 初始化配置
	initConfig()

	// 2. 连接数据库
	initDB()

	// 3. 自动迁移表结构
	if err := db.AutoMigrate(&model.TrainingJob{}); err != nil {
		log.Fatalf("数据库表迁移失败: %v", err)
	}

	// 4. 初始化依赖组件
	jobRep := repository.NewJobRepository(db)
	k8sClient, err := k8s.NewK8sClient()
	if err != nil {
		log.Fatalf("K8s客户端连接失败: %s", err)
	}
	log.Println("K8s客户端连接成功")
	jobSvc := service.NewJobService(jobRep, k8sClient)
	jobHandler := handler.NewJobHandler(jobSvc)

	// 5. 创建上下文（用于控制所有协程优雅退出）
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 6. 启动【状态同步协程】（优雅启停版）
	go syncJobStatus(ctx, jobSvc)

	// 7. 初始化Gin路由
	r := gin.Default()
	api := r.Group("/api/v1")
	{
		// jobs
		api.POST("/jobs", jobHandler.Create)          // 提交训练任务
		api.GET("/jobs", jobHandler.List)             // 列表任务
		api.GET("/jobs/:id", jobHandler.Get)          // 获取任务详情
		api.DELETE("/jobs/:id", jobHandler.Delete)    // 删除任务 (停止Job并删除记录)
		api.GET("/jobs/:id/logs", jobHandler.GetLogs) // 获取任务日志
	}

	// 8. 启动HTTP服务（支持优雅关闭）
	servicePort := viper.GetString("server.jobPort")
	if servicePort == "" {
		servicePort = "58081"
	}
	srv := &http.Server{
		Addr:    ":" + servicePort,
		Handler: r,
	}
	// 9. 异步启动HTTP服务
	go func() {
		log.Printf("Job服务启动成功，监听端口: %s\n", servicePort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP服务启动失败: %v", err)
		}
	}()

	// 9. 监听系统退出信号（优雅启停核心）
	quit := make(chan os.Signal, 1)
	// 监听：Ctrl+C / kill 命令
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // 阻塞等待退出信号
	log.Println("收到退出信号，开始优雅关闭服务...")

	// 10. 执行优雅关闭流程
	// 1. 取消上下文，停止状态同步协程
	cancel()
	// 2. 关闭HTTP服务（等待5秒内处理完现有请求）
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP服务关闭失败: %v", err)
	}
	// 3. 关闭数据库连接池
	sqlDB, err := db.DB()
	if err == nil {
		_ = sqlDB.Close()
		log.Println("数据库连接已关闭")
	}

	log.Println("服务已完全退出，欢迎再次使用！")

}
