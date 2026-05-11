package handler

import (
	"github.com/gin-gonic/gin"
	"kubeai/internal/jobscheduler/req"
	"kubeai/internal/jobscheduler/service"
	"net/http"
	"strconv"
)

type JobHandler struct {
	svc *service.JobService
}

func NewJobHandler(svc *service.JobService) *JobHandler {
	return &JobHandler{
		svc: svc,
	}
}

func (h *JobHandler) Create(c *gin.Context) {
	var req req.CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	job, err := h.svc.CreateJob(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, job)
}

func (h *JobHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	job, err := h.svc.GetJob(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "jobscheduler not found"})
		return
	}
	c.JSON(http.StatusOK, job)
}

func (h *JobHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	status := c.Query("status")
	jobs, total, err := h.svc.ListJobs(page, size, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": status,
		"items":  jobs,
		"total":  total,
	})
}

func (h *JobHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.DeleteJob(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *JobHandler) GetLogs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	tail := c.Query("tail")
	var tailLines int64
	if tail != "" {
		tailLines, _ = strconv.ParseInt(tail, 10, 64)
	}
	log, err := h.svc.GetJobLog(c.Request.Context(), uint(id), tailLines)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, log)
}
