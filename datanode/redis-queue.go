package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mrowaha/dos/datanode"
	"github.com/mrowaha/dos/namenode"
	"github.com/redis/go-redis/v9"
)

const (
	createListKey = "create-objects"
	deliveryQueue = "delivery-queue"
)

type RedisDataNodeQueue struct {
	redisClient *redis.Client
}

func NewRedisDataNodeQueue() *RedisDataNodeQueue {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Update with your Redis server address
		DB:   0,
	})
	keys, err := redisClient.Keys(context.TODO(), fmt.Sprintf("%s:%s:*", process, createListKey)).Result()
	if err != nil {
		panic(err)
	}

	// Delete the keys
	if len(keys) > 0 {
		_, err := redisClient.Del(context.TODO(), keys...).Result()
		if err != nil {
			panic(err)
		}
		fmt.Printf("Deleted %d queue keys\n", len(keys))
	} else {
		fmt.Println("no stale create object packets")
	}
	_, err = redisClient.Del(context.TODO(), fmt.Sprintf("%s:%s", process, deliveryQueue)).Result()
	if err != nil {
		panic(err)
	}
	return &RedisDataNodeQueue{
		redisClient,
	}
}

func (r *RedisDataNodeQueue) PushCreateCmd(cmd namenode.CreateCommand) error {
	data, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed to marshal CreateCommand: %w", err)
	}

	ctx := context.Background()
	queue := fmt.Sprintf("%s:%s:%s", process, createListKey, cmd.Name)
	// Push serialized data to the tail of the Redis list
	err = r.redisClient.RPush(ctx, queue, data).Err()
	if err != nil {
		return fmt.Errorf("failed to push CreateCommand to Redis list: %w", err)
	}
	fmt.Printf("pushed packet %v to queue %s\n", cmd, queue)
	return nil
}

func (r *RedisDataNodeQueue) PullCreateCmd(object string) (*namenode.CreateCommand, error) {
	queue := fmt.Sprintf("%s:%s:%s", process, createListKey, object)
	fmt.Printf("pulling packet from %s queue\n", queue)

	// Pop data from the head of the Redis list
	data, err := r.redisClient.LPop(context.Background(), queue).Result()
	if err != nil {
		if err == redis.Nil {
			// If the list is empty, return nil without an error
			fmt.Printf("no packets left in queue %s\n", queue)
			return nil, nil
		}
		return nil, fmt.Errorf("failed to pull CreateCommand from Redis list: %w", err)
	}

	var cmd namenode.CreateCommand
	// Deserialize the data
	if err := json.Unmarshal([]byte(data), &cmd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CreateCommand: %w", err)
	}

	return &cmd, nil
}
func (r *RedisDataNodeQueue) BlockCommand(score float64, cmd interface{}) error {
	data, _ := json.Marshal(cmd)
	err := r.redisClient.ZAdd(context.TODO(), fmt.Sprintf("%s:%s", process, deliveryQueue), redis.Z{
		Score:  score,
		Member: data,
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to insert to min-heap: %w", err)
	}
	return nil
}

func (r *RedisDataNodeQueue) RetrieveCommand() (interface{}, float64, error) {
	result, err := r.redisClient.ZRangeWithScores(context.TODO(), fmt.Sprintf("%s:%s", process, deliveryQueue), 0, 0).Result()
	if err != nil {
		return "", 0, fmt.Errorf("failed to get min from min-heap: %w", err)
	}
	if len(result) == 0 {
		return "", 0, datanode.ErrNoCommandToRetrieve // Min-heap is empty
	}
	var cmd map[string]any
	json.Unmarshal([]byte(result[0].Member.(string)), &cmd)
	return cmd, result[0].Score, nil
}

func (r *RedisDataNodeQueue) DeliverCommand() (interface{}, float64, error) {
	result, err := r.redisClient.ZPopMin(context.TODO(), fmt.Sprintf("%s:%s", process, deliveryQueue), 1).Result()
	if err != nil {
		return "", 0, fmt.Errorf("failed to remove min from min-heap: %w", err)
	}
	if len(result) == 0 {
		return "", 0, nil // Min-heap is empty
	}

	var cmd map[string]any
	json.Unmarshal([]byte(result[0].Member.(string)), &cmd)
	_eventType, ok := cmd["type"]
	if !ok {
		return "", 0, errors.New("internal error. should have type prop on blocked command")
	}

	eventType := namenode.BroadcastEvent(_eventType.(string))
	switch eventType {
	case namenode.COMMIT:
		return namenode.CommitCommand{
			Type: eventType,
		}, result[0].Score, nil
	case namenode.DELETE:
		return namenode.DeleteCommand{
			Name: cmd["name"].(string),
			Type: eventType,
		}, result[0].Score, nil
	case namenode.UPDATE:
		return namenode.UpdateCommand{
			Name: cmd["name"].(string),
			Data: []byte(cmd["data"].(string)),
			Type: eventType,
		}, result[0].Score, nil
	default:
		return "", 0, fmt.Errorf("failed to deliver, unexpected event type %s\n", eventType)
	}

}
