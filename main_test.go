package main_test

import (
	"context"
	"testing"
	"time"

	"github.com/redis/rueidis"
)

func TestXxx(t *testing.T) {
	rds, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{"localhost:6379"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer rds.Close()

	if err := rds.Do(context.Background(), rds.B().Ping().Build()).Error(); err != nil {
		t.Fatal(err)
	}

	setCmd := rds.B().Set().Key("key").Value("value").Build()

	if err := rds.Do(context.Background(), setCmd).Error(); err != nil {
		t.Fatal(err)
	}

	getCmd := rds.B().Get().Key("key").Build()

	v, err := rds.Do(context.Background(), getCmd).ToString()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(v)

	cacheCmd1 := rds.B().Get().Key("key").Cache()

	v, err = rds.DoCache(context.Background(), cacheCmd1, time.Second).ToString()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(v)

	cacheCmd2 := rds.B().Get().Key("key").Cache()

	v, err = rds.DoCache(context.Background(), cacheCmd2, time.Second).ToString()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(v)
}
