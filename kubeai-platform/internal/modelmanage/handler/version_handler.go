package handler

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/minio/minio-go/v7"
	"gorm.io/datatypes"
	"kubeai/internal/model"
	service2 "kubeai/internal/modelmanage/service"
	"kubeai/internal/pkg/storage"
	"net/http"
	"strconv"
)

type VersionHandler struct {
	svc         *service2.VersionService
	m           *service2.ModelService
	minioClient *storage.MinIOClient
}

func NewVersionHandler(svc *service2.VersionService,
	m *service2.ModelService,
	minioClient *storage.MinIOClient) *VersionHandler {
	return &VersionHandler{
		svc:         svc,
		m:           m,
		minioClient: minioClient,
	}
}

func (h *VersionHandler) Get(c *gin.Context) {
	modelID, err := strconv.ParseUint(c.Param("modelID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	model, err := h.m.GetByID(uint(modelID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
	}
	version := c.Param("version")
	mv, err := h.svc.GetByModelIDAndVersion(model.ID, version)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "modelVersion not found"})
	}
	c.JSON(http.StatusOK, gin.H{"data": mv})
}

func (h *VersionHandler) GetMetadata(c *gin.Context) {
	modelID, err := strconv.ParseUint(c.Param("modelID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m, err := h.m.GetByID(uint(modelID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
	}
	version := c.Param("version")
	mv, err := h.svc.GetByModelIDAndVersion(m.ID, version)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "modelVersion not found"})
	}
	metadata := &model.Metadata{
		ModelName:    m.Name,
		ModelVersion: mv.ModelVersion,
		StoragePath:  mv.StoragePath,
	}
	c.JSON(http.StatusOK, gin.H{"data": metadata})
}

func (h *VersionHandler) List(c *gin.Context) {
	modelID, err := strconv.ParseUint(c.Param("modelID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	model, err := h.m.GetByID(uint(modelID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
	}
	mvs, total, err := h.svc.GetByModelID(model.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": mvs, "total": total})
}

func (h *VersionHandler) Upload(c *gin.Context) {
	modelID, err := strconv.ParseUint(c.Param("modelID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 1、解析表单
	version := c.PostForm("version")
	framework := c.PostForm("framework")
	metricStr := c.PostForm("metrics")
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	// 2、验证model 存在
	m, err := h.m.GetByID(uint(modelID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
		return
	}

	// 3.生成存储路径
	storagePath := fmt.Sprintf("models/%s/%s/model.bin", m.Name, version)

	// 4. 上传到MinIO
	ctx := context.Background()
	if err := h.minioClient.UploadFile(ctx, storagePath, file, header.Size, header.Header.Get("Content-Type")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
		return
	}

	// 5. 创建版本记录
	var metrics datatypes.JSON
	if metricStr != "" {
		metrics = datatypes.JSON([]byte(metricStr))
	}
	v := model.ModelVersion{
		ModelID:      uint(modelID),
		ModelVersion: version,
		StoragePath:  storagePath,
		Framework:    framework,
		Metrics:      metrics,
	}

	if _, err := h.svc.Create(v); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": v})
}

func (h *VersionHandler) UpdateStatus(c *gin.Context) {
	modelID, err := strconv.ParseUint(c.Param("modelID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	version := c.Param("version")
	_, err = h.m.GetByID(uint(modelID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
		return
	}
	mv, err := h.svc.GetByModelIDAndVersion(uint(modelID), version)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "modelVersion not found"})
		return
	}
	mv.Status = c.Param("status")
	result, err := h.svc.Update(mv)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *VersionHandler) Download(c *gin.Context) {
	modelID, err := strconv.ParseUint(c.Param("modelID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err = h.m.GetByID(uint(modelID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
		return
	}
	version := c.Param("version")
	mv, err := h.svc.GetByModelIDAndVersion(uint(modelID), version)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "modelVersion not found"})
		return
	}
	file, err := h.minioClient.DownloadFile(context.Background(), mv.StoragePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": file})
}
