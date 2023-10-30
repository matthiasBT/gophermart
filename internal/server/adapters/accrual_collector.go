package adapters

import (
	"context"
	"time"

	"github.com/matthiasBT/gophermart/internal/infra/logging"
	"github.com/matthiasBT/gophermart/internal/server/entities"
)

type Collector struct {
	name        string
	client      entities.IAccrualClient
	storage     entities.Storage
	orderRepo   entities.OrderRepo
	accrualRepo entities.AccrualRepo
	logger      logging.ILogger
	jobs        <-chan entities.Job
	done        <-chan struct{}
}

func NewCollector(
	name string,
	client entities.IAccrualClient,
	storage entities.Storage,
	orderRepo entities.OrderRepo,
	accrualRepo entities.AccrualRepo,
	logger logging.ILogger,
	jobs <-chan entities.Job,
	done <-chan struct{},
) *Collector {
	return &Collector{
		name:        name,
		client:      client,
		storage:     storage,
		orderRepo:   orderRepo,
		accrualRepo: accrualRepo,
		logger:      logger,
		jobs:        jobs,
		done:        done,
	}
}

func (c *Collector) Run(ctx context.Context) {
	c.logger.Infof("Launching the Collector worker %s", c.name)
	for {
		select {
		case <-c.done:
			c.logger.Infof("Stopping the Collector worker %s", c.name)
			return
		case job := <-c.jobs:
			c.logger.Infof("New job for collector %s: %v", c.name, job)
			if err := c.collect(ctx, &job); err != nil {
				c.logger.Errorf("Collector %s job %s failed: %v", c.name, job, err)
			}
		}
	}
}

func (c *Collector) collect(ctx context.Context, job *entities.Job) error {
	if order, err := c.orderRepo.FindOrder(ctx, job.OrderNumber); err != nil {
		return err
	} else if order.Status == "INVALID" || order.Status == "PROCESSED" {
		c.logger.Infof("Order %s was already fetched from the accrual service")
		return nil
	}
	resp, err := c.client.GetAccrual(ctx, job.OrderNumber)
	if err != nil {
		return err
	}
	tx, err := c.storage.Tx(ctx)
	if err != nil {
		return err
	}
	defer tx.Commit()
	if err := c.accrualRepo.CreateAccrual(ctx, tx, job.UserID, resp); err != nil {
		defer tx.Rollback()
		return err
	}
	if err := c.orderRepo.UpdateOrderStatus(ctx, tx, job.OrderNumber, resp.Status); err != nil {
		defer tx.Rollback()
		return err
	}
	return nil
}

type Supplier struct {
	storage   entities.Storage
	orderRepo entities.OrderRepo
	logger    logging.ILogger
	jobs      chan<- entities.Job
	done      <-chan struct{}
	tick      <-chan time.Time
	batchSize int
}

func NewSupplier(
	storage entities.Storage,
	orderRepo entities.OrderRepo,
	logger logging.ILogger,
	jobs chan<- entities.Job,
	done <-chan struct{},
	tick <-chan time.Time,
	batchSize int,
) *Supplier {
	return &Supplier{
		storage:   storage,
		orderRepo: orderRepo,
		logger:    logger,
		jobs:      jobs,
		done:      done,
		tick:      tick,
		batchSize: batchSize,
	}
}

func (s *Supplier) Run(ctx context.Context) {
	for {
		select {
		case <-s.done:
			s.logger.Infoln("Stopping the Supplier worker")
			return
		case tick := <-s.tick:
			s.logger.Infof("Supplier worker is ticking at %v", tick)
			if err := s.supply(ctx); err != nil {
				s.logger.Errorf("Supplier worker failed: %v", err)
			}
		}
	}
}

func (s *Supplier) supply(ctx context.Context) error {
	orders, err := s.orderRepo.FetchUnprocessedOrders(ctx, s.batchSize)
	if err != nil {
		return err
	}
	if orders == nil {
		s.logger.Infoln("No orders to process")
		return nil
	}
	s.logger.Infof("Fetched %d orders for processing", len(orders))
	for _, order := range orders {
		s.jobs <- entities.Job{
			UserID:      order.UserID,
			OrderNumber: order.Number,
		}
	}
	s.logger.Infoln("Orders scheduled")
	return nil
}
