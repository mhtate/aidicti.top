package discovery

import (
	"context"
	"errors"
	"time"

	"aidicti.top/pkg/logging"
	"aidicti.top/pkg/utils"
)

// Registry defines a service registry.
type Registry interface {
	// Register creates a service instance record in the registry.
	Register(ctx context.Context, instanceID string, serviceName string, hostPort string) error
	// Deregister removes a service insttance record from the registry.
	Deregister(ctx context.Context, instanceID string, serviceName string) error
	// ServiceAddresses returns the list of addresses of active instances of the given service.
	ServiceAddresses(ctx context.Context, serviceID string) ([]string, error)
	// ReportHealthyState is a push mechanism for reporting healthy state to the registry.
	ReportHealthyState(instanceID string, serviceName string) error
}

// ErrNotFound is returned when no service addresses are found.
var ErrNotFound = errors.New("no service addresses found")

func RegisterService(ctx context.Context, serviceName string, instanceID string, hostPort string, reg Registry) (func(), string, error) {
	err := reg.Register(ctx, instanceID, serviceName, hostPort)
	utils.Assert(err == nil, "register service in consul failed")

	cancelCtx, cancelF := context.WithCancel(ctx)

	healthCheckTimeout := 4 * time.Second

	reportHealth := func() {
		err := reg.ReportHealthyState(instanceID, serviceName)
		if err != nil {
			logging.Info("report healthy state failed", "err", err)
			return
		}

		// logging.Debug("report healthy state finished")
	}

	reportHealth()

	go func() {
		for {
			select {
			case <-time.After(healthCheckTimeout):
				reportHealth()

			case <-cancelCtx.Done():
				logging.Info("report healthy state finished by context")
				return
			}
		}
	}()

	return func() {
		logging.Info("deregister service in consul started")

		cancelF()

		err := reg.Deregister(cancelCtx, instanceID, serviceName)
		if err != nil {
			logging.Info("deregister service in consul failed", "err", err)
			return
		}

		logging.Info("deregister service in consul finished")
	}, instanceID, nil
}
