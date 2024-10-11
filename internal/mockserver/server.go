package mockserver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func Serve(c context.Context, port string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/", Api)
	mux.HandleFunc("/RPC2_Login", Rpc2Login)
	mux.HandleFunc("/login.php", HeadlessLogin)
	mux.HandleFunc("/headless", Headless)
	mux.HandleFunc("/ISAPI/Security/sessionLogin/capabilities", CVE_2020_7057)

	serve(c, fmt.Sprintf(":%s", port), mux)
}

func serve(c context.Context, addr string, handler http.Handler) {
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	logger := log.New(os.Stdout, "[MockServer] ", log.LstdFlags|log.Lshortfile)
	go func() {
		logger.Printf("server listen on %s...", addr)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Printf("listen: %s\n", err)
		}
	}()

	<-c.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("server force to shutdown: %s\n", err)
	}

	logger.Println("server exiting")
}
