package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

// TaskQueue 基于 Redis Stream 实现任务队列，解耦调度与执行，支持异步任务处理
type TaskQueue struct {
	redisClient *redis.Client
	streamKey   string
	group       string
	consumer    string
}

const (
	InferenceTaskStream = "kubeai:inference:tasks"
	TrainingTaskStream  = "kubeai:training:tasks"
	ConsumerGroup       = "kubeai-scheduler"
)

func NewTaskQueue(redisClient *redis.Client, streamKey string) *TaskQueue {
	return &TaskQueue{
		redisClient: redisClient,
		streamKey:   streamKey,
		group:       ConsumerGroup,
		consumer:    fmt.Sprintf("%s-%d", ConsumerGroup, time.Now().Unix()),
	}
}

// Init 初始化消费者组
func (q *TaskQueue) Init(ctx context.Context) error {
	_, err := q.redisClient.XGroupCreateMkStream(ctx, q.streamKey, q.group, "0").Result()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return err
	}
	return nil
}

// PushTask 下发任务到队列
func (q *TaskQueue) PushTask(ctx context.Context, task interface{}) error {
	taskData, err := json.Marshal(task)
	if err != nil {
		return err
	}
	_, err = q.redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: q.streamKey,
		Values: map[string]interface{}{
			"task": taskData,
			"time": time.Now().Unix(),
		},
	}).Result()
	return err
}

// ConsumeTask 消费任务
func (q *TaskQueue) ConsumeTask(ctx context.Context, handler func(taskData string) error) {
	for {
		// 读取新消息
		entries, err := q.redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    q.group,
			Consumer: q.consumer,
			Streams:  []string{q.streamKey, ">"},
			Count:    1,
			Block:    5 * time.Second,
		}).Result()
		if err != nil || len(entries) == 0 {
			continue
		}
		// 处理消息
		for _, entry := range entries[0].Messages {
			if err := handler(entry.Values["task"].(string)); err == nil {
				// 处理成功，确认消息
				q.redisClient.XAck(ctx, q.streamKey, q.group, entry.ID).Result()
			}
		}
	}

}
