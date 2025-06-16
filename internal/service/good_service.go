package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"go-test/internal/logger"
	"go-test/internal/model"
	"go-test/internal/repo"
	"time"

	"github.com/redis/go-redis/v9"
)

type GoodService interface {
	Create(ctx context.Context, g *model.Good) error
	GetByID(ctx context.Context, id int) (*model.Good, error)
	Update(ctx context.Context, g *model.Good) error
	Delete(ctx context.Context, id int, projectID int) (*model.Good, error)
	List(ctx context.Context, projectID, limit, offset int, sort string) ([]model.Good, int, int, error)
	Reprioritize(ctx context.Context, id, projectID, newPriority int) ([]model.Good, error)
}

type goodService struct {
	repo   repo.GoodRepository
	redis  *redis.Client
	logger logger.Logger
}

func NewGoodService(r repo.GoodRepository, redis *redis.Client, logger logger.Logger) *goodService {
	return &goodService{
		repo:   r,
		redis:  redis,
		logger: logger,
	}
}

func (s *goodService) Create(ctx context.Context, g *model.Good) error {
	max, err := s.repo.GetMaxPriority(ctx, g.ProjectID)
	if err != nil {
		return err
	} else {
		g.Priority = max + 1
	}

	g.CreatedAt = time.Now()

	if g.Name == "" {
		return errors.New("validation error: name is required")
	}

	err = s.repo.Create(ctx, g)
	if err != nil {
		return err
	}

	_ = s.logger.Publish(logger.Event{
		ID:        g.ID,
		ProjectID: g.ProjectID,
		Action:    "created",
		Timestamp: time.Now(),
	})

	return nil
}

func (s *goodService) GetByID(ctx context.Context, id int) (*model.Good, error) {
	g, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("good not found: %w", err)
		}
		return nil, err
	}

	return g, nil
}

func (s *goodService) Update(ctx context.Context, g *model.Good) error {
	if g.Name == "" {
		return errors.New("validation error: name is required")
	}

	err := s.repo.Update(ctx, g)
	if err != nil {
		return err
	}

	_ = s.logger.Publish(logger.Event{
		ID:        g.ID,
		ProjectID: g.ProjectID,
		Action:    "updated",
		Timestamp: time.Now(),
	})

	s.invalidateGoodsCache(ctx, g.ProjectID)
	return nil
}

func (s *goodService) Delete(ctx context.Context, id int, projectID int) (*model.Good, error) {
	g, err := s.repo.Delete(ctx, id, projectID)
	if err != nil {
		return nil, err
	}

	_ = s.logger.Publish(logger.Event{
		ID:        g.ID,
		ProjectID: g.ProjectID,
		Action:    "deleted",
		Timestamp: time.Now(),
	})

	s.invalidateGoodsCache(ctx, projectID)
	return g, nil
}

func (s *goodService) List(ctx context.Context, projectID, limit, offset int, sort string) ([]model.Good, int, int, error) {
	cacheKey := fmt.Sprintf("goods:project=%d:limit=%d:offset=%d:sort=%s", projectID, limit, offset, sort)

	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var cachedResult struct {
			Goods        []model.Good `json:"goods"`
			TotalCount   int          `json:"total"`
			RemovedCount int          `json:"removed"`
		}
		if err := json.Unmarshal([]byte(cached), &cachedResult); err == nil {
			return cachedResult.Goods, cachedResult.TotalCount, cachedResult.RemovedCount, nil
		}
	}
	goods, total, removed, err := s.repo.List(ctx, projectID, limit, offset, sort)
	if err != nil {
		return nil, 0, 0, err
	}

	cachedData := struct {
		Goods        []model.Good `json:"goods"`
		TotalCount   int          `json:"total"`
		RemovedCount int          `json:"removed"`
	}{
		Goods:        goods,
		TotalCount:   total,
		RemovedCount: removed,
	}

	bytes, err := json.Marshal(cachedData)
	if err == nil {
		_ = s.redis.Set(ctx, cacheKey, bytes, time.Minute).Err()
	}

	return goods, total, removed, nil
}

func (s *goodService) Reprioritize(ctx context.Context, id, projectID, newPriority int) ([]model.Good, error) {
	goods, err := s.repo.Reprioritize(ctx, id, projectID, newPriority)
	if err != nil {
		return nil, err
	}

	_ = s.logger.Publish(logger.Event{
		ID:        id,
		ProjectID: projectID,
		Action:    "reprioritized",
		Timestamp: time.Now(),
	})

	s.invalidateGoodsCache(ctx, projectID)
	return goods, nil
}

func (s *goodService) invalidateGoodsCache(ctx context.Context, projectID int) {
	pattern := fmt.Sprintf("goods:project=%d:*", projectID)
	iter := s.redis.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		_ = s.redis.Del(ctx, iter.Val()).Err()
	}
}
