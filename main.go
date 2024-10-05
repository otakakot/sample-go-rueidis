package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/redis/rueidis"

	"github.com/otakakot/sample-go-rueidis/internal/cache"
)

func main() {
	port := cmp.Or(os.Getenv("PORT"), "8080")

	address := cmp.Or(os.Getenv("REDIS_ADDRESS"), "redis:6379")

	mode := cmp.Or(os.Getenv("MODE"), "rueidis")

	mux := http.NewServeMux()

	var cach cache.Cache[string]

	if mode == "inmemory" {
		rds := redis.NewClient(&redis.Options{
			Addr: address,
		})
		defer rds.Close()

		if err := rds.Ping(context.Background()).Err(); err != nil {
			panic(err)
		}

		cach = cache.NewInMemory[string](rds)
	} else {
		rds, err := rueidis.NewClient(rueidis.ClientOption{
			InitAddress: []string{address},
		})
		if err != nil {
			panic(err)
		}
		defer rds.Close()

		if err := rds.Do(context.Background(), rds.B().Ping().Build()).Error(); err != nil {
			panic(err)
		}

		cach = cache.NewRueidis[string](rds)
	}

	hdl := &Handler{
		cach:  cach,
		value: port,
	}

	mux.HandleFunc("/", hdl.Handle)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", port),
		Handler:           mux,
		ReadHeaderTimeout: 30 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	defer stop()

	go func() {
		slog.Info("start server listen port: " + port)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	<-ctx.Done()

	slog.Info("start server shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		panic(err)
	}

	slog.Info("done server shutdown")
}

type Handler struct {
	cach  cache.Cache[string]
	value string
}

func (hdl *Handler) Handle(
	rw http.ResponseWriter,
	req *http.Request,
) {
	switch req.Method {
	case http.MethodGet:
		hdl.get(rw, req)

		return
	case http.MethodPost:
		hdl.post(rw, req)

		return
	case http.MethodPut:
		hdl.put(rw, req)

		return
	case http.MethodDelete:
		hdl.delete(rw, req)

		return
	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (hdl *Handler) get(
	rw http.ResponseWriter,
	req *http.Request,
) {
	val, err := hdl.cach.Get(req.Context(), "key")
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			rw.WriteHeader(http.StatusNotFound)

			rw.Write([]byte(err.Error() + "\n"))

			return
		} else {
			rw.WriteHeader(http.StatusInternalServerError)

			rw.Write([]byte(err.Error() + "\n"))

			return
		}
	}

	rw.Write([]byte(*val + "\n"))
}

func (hdl *Handler) post(
	rw http.ResponseWriter,
	req *http.Request,
) {
	if err := hdl.cach.Set(req.Context(), "key", "value"); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)

		rw.Write([]byte(err.Error() + "\n"))
	}

	rw.WriteHeader(http.StatusCreated)

	rw.Write([]byte("Created\n"))
}

func (hdl *Handler) put(
	rw http.ResponseWriter,
	req *http.Request,
) {
	if err := hdl.cach.Set(req.Context(), "key", hdl.value); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)

		rw.Write([]byte(err.Error() + "\n"))
	}

	rw.WriteHeader(http.StatusCreated)

	rw.Write([]byte("Created" + "\n"))
}

func (hdl *Handler) delete(
	rw http.ResponseWriter,
	req *http.Request,
) {
	if err := hdl.cach.Del(req.Context(), "key"); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)

		rw.Write([]byte(err.Error() + "\n"))

		return
	}

	rw.WriteHeader(http.StatusNoContent)

	rw.Write([]byte("No Content" + "\n"))
}
