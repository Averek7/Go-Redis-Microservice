package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/averek7/order-api/model"
	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	Client *redis.Client
}

func OrderIdKey(orderId uint64) string {
	return fmt.Sprintf("order:%d", orderId)
}

func (r *RedisRepo) Insert(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order) // convert order to json
	if err != nil {
		return fmt.Errorf("Failed to marshal order: %w", err)
	}

	key := OrderIdKey(order.OrderID) // create key for redis

	txn := r.Client.TxPipeline() // create transaction

	res := txn.SetNX(ctx, key, string(data), 0) // insert order into redis
	if err := res.Err(); err != nil {
		return fmt.Errorf("Failed to insert order: %w", err)
	}
	if err := txn.SAdd(ctx, "orders", key).Err(); err != nil { // add order to set
		return fmt.Errorf("Failed to add order to set: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil { // execute transaction
		return fmt.Errorf("Failed to execute transaction: %w", err)
	}

	return nil
}

var ErrNotExist = errors.New("order does not exist")

func (r *RedisRepo) FindById(ctx context.Context, orderId uint64) (model.Order, error) {
	key := OrderIdKey(orderId)                   // create key for redis
	data, err := r.Client.Get(ctx, key).Result() // get order from redis

	if errors.Is(err, redis.Nil) {
		return model.Order{}, ErrNotExist
	} else if err != nil {
		return model.Order{}, fmt.Errorf("Failed to get order: %w", err)
	}

	var order model.Order
	err = json.Unmarshal([]byte(data), &order) // convert json to order

	if err != nil {
		return model.Order{}, fmt.Errorf("Failed to unmarshal order: %w", err)
	}

	return order, nil
}

func (r *RedisRepo) DeleteById(ctx context.Context, orderId uint64) error {
	key := OrderIdKey(orderId) // create key for redis

	txn := r.Client.TxPipeline() // create transaction

	err := txn.Del(ctx, key).Err() // delete order from redis
	if errors.Is(err, redis.Nil) {
		txn.Discard()
		return ErrNotExist
	} else if err != nil {
		txn.Discard()
		return fmt.Errorf("Failed to get order: %w", err)
	}

	if err := txn.SRem(ctx, "orders", key).Err(); err != nil { // remove order from set
		txn.Discard()
		return fmt.Errorf("Failed to remove order from set: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil { // execute transaction
		return fmt.Errorf("Failed to execute transaction: %w", err)
	}
	return nil
}

func (r *RedisRepo) Update(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order) // convert order to json
	if err != nil {
		return fmt.Errorf("Failed to marshal order: %w", err)
	}

	key := OrderIdKey(order.OrderID)                      // create key for redis
	err = r.Client.SetXX(ctx, key, string(data), 0).Err() //  update order in redis

	if errors.Is(err, redis.Nil) {
		return ErrNotExist
	} else if err != nil {
		return fmt.Errorf("Failed to insert order: %w", err)
	}

	return nil
}

type FindAllPage struct {
	Size   uint
	Offset uint
}

type FindResult struct {
	Orders []model.Order
	Cursor uint64
}

func (r *RedisRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	res := r.Client.SScan(ctx, "orders", uint64(page.Offset), "", int64(page.Size)) // get orders from set

	key, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("Failed to get orders: %w", err)
	}

	if len(key) == 0 {
		return FindResult{
			Orders: []model.Order{},
		}, nil
	}

	xs, err := r.Client.MGet(ctx, key...).Result() // get orders from redis
	if err != nil {
		return FindResult{}, fmt.Errorf("Failed to get orders: %w", err)
	}

	orders := make([]model.Order, len(xs))

	for i, x := range xs {
		x := x.(string)
		var order model.Order

		err = json.Unmarshal([]byte(x), &order) // convert json to order
		if err != nil {
			return FindResult{}, fmt.Errorf("Failed to unmarshal order: %w", err)
		}
		orders[i] = order
	}

	return FindResult{
		Orders: orders,
		Cursor: cursor,
	}, nil
}
