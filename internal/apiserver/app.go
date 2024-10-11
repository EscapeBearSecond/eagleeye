package apiserver

import (
	"context"
	"errors"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

func Run(c context.Context, servers ...*HTTPServer) error {
	Logger.Info("APIServer Startup")

	egRun := errgroup.Group{}

	for _, server := range servers {
		egRun.Go(func() error {
			Logger.Info("Server Listen", "server_id", server.ID, "addr", server.Addr)
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				return err
			}
			return nil
		})
	}

	<-c.Done()

	cc, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	egShutdown := errgroup.Group{}
	for _, server := range servers {
		egShutdown.Go(func() error {
			Logger.Info("Server Shutdown", "server_id", server.ID, "addr", server.Addr)
			if err := server.Shutdown(cc); err != nil {
				return err
			}
			return nil
		})
	}

	if err := egShutdown.Wait(); err != nil {
		return err
	}

	if err := egRun.Wait(); err != nil {
		return err
	}
	Logger.Info("APIServer Exit")
	return nil
}
