package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"kubeai/internal/model"
	handler2 "kubeai/internal/modelmanage/handler"
	repository2 "kubeai/internal/modelmanage/repository"
	service2 "kubeai/internal/modelmanage/service"
	"kubeai/internal/pkg/storage"
	"log"
)

func initConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath("../../configs")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
}

func main() {
	initConfig()

	// 连接数据库
	host := viper.GetString("database.host")
	port := viper.GetInt("database.port")
	user := viper.GetString("database.user")
	password := viper.GetString("database.password")
	dbname := viper.GetString("database.dbname")
	sslmode := viper.GetString("database.sslmode")
	dsn := fmt.Sprintf("host=%s:%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to database, %s", err)
	}

	// 自动迁移
	db.AutoMigrate(&model.Model{}, &model.ModelVersion{})

	// 初始化 MinIO
	minIOClient, err := storage.NewMinIOClient(
		viper.GetString("minio.endpoint"),
		viper.GetString("minio.accessKey"),
		viper.GetString("minio.secretKey"),
		viper.GetString("minio.bucket"),
		viper.GetBool("minio.useSSL"),
	)
	if err != nil {
		log.Fatalf("Error connecting to minIO client, %s", err)
	}

	// 初始化 repository
	modelRep := repository2.NewModelRepository(db)
	versionRep := repository2.NewVersionRepository(db)

	// 初始化 service
	modelSvc := service2.NewModelService(modelRep)
	versionSvc := service2.NewVersionService(versionRep)

	// 初始化 handler
	modelHandler := handler2.NewModelHandler(modelSvc)
	versionHandler := handler2.NewVersionHandler(versionSvc, modelSvc, minIOClient)

	// 路由
	r := gin.Default()
	api := r.Group("/api/v1")
	{
		// models
		api.POST("/models", modelHandler.Create)
		api.GET("/models", modelHandler.List)
		api.GET("/models/:id", modelHandler.Get)
		api.PUT("/models/:id", modelHandler.Update)
		api.DELETE("/models/:id", modelHandler.Delete)

		// versions
		api.POST("/models/:modelId/versions", versionHandler.Upload)
		api.GET("/models/:modelId/versions", versionHandler.List)
		api.GET("/models/:modelId/versions/:version", versionHandler.Get)
		api.GET("/models/:modelId/versions/:version/metadata", versionHandler.GetMetadata)
		api.PUT("/models/:modelId/versions/:version/status", versionHandler.UpdateStatus)
		api.GET("/models/:modelId/versions/:version/download", versionHandler.Download)
	}

	servicePort := viper.GetString("server.modelPort")
	if servicePort == "" {
		servicePort = "58080"
	}
	r.Run(":" + servicePort)
}
