package job_test

import (
	"context"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cozy/cozy-stack/model/instance"
	"github.com/cozy/cozy-stack/model/job"
	"github.com/cozy/cozy-stack/pkg/config/config"
	"github.com/cozy/cozy-stack/pkg/limits"
	"github.com/cozy/cozy-stack/tests/testutils"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

const redisURL1 = "redis://localhost:6379/0"
const redisURL2 = "redis://localhost:6379/1"

func TestRedisBroker(t *testing.T) {
	if testing.Short() {
		t.Skip("an instance is required for this test: test skipped due to the use of --short flag")
	}

	config.UseTestFile(t)
	setup := testutils.NewSetup(t, t.Name())
	testInstance := setup.GetTestInstance()

	t.Run("RedisJobs", func(t *testing.T) {
		job.SetRedisTimeoutForTest()
		opts1, _ := redis.ParseURL(redisURL1)
		opts2, _ := redis.ParseURL(redisURL2)
		client1 := redis.NewClient(opts1)
		client2 := redis.NewClient(opts2)

		n := 10
		v := 100

		var w sync.WaitGroup
		w.Add(2*n + 1)

		workersTestList := job.WorkersList{
			{
				WorkerType:  "test",
				Concurrency: 4,
				WorkerFunc: func(ctx *job.WorkerContext) error {
					var msg string
					err := ctx.UnmarshalMessage(&msg)
					if !assert.NoError(t, err) {
						return err
					}
					if strings.HasPrefix(msg, "z-") {
						_, err := strconv.Atoi(msg[len("z-"):])
						assert.NoError(t, err)
					} else if strings.HasPrefix(msg, "a-") {
						_, err := strconv.Atoi(msg[len("a-"):])
						assert.NoError(t, err)
					} else if strings.HasPrefix(msg, "b-") {
						_, err := strconv.Atoi(msg[len("b-"):])
						assert.NoError(t, err)
					} else {
						t.Fatal()
					}
					w.Done()
					return nil
				},
			},
		}

		broker1 := job.NewRedisBroker(client1)
		err := broker1.StartWorkers(workersTestList)
		assert.NoError(t, err)

		broker2 := job.NewRedisBroker(client2)
		err = broker2.StartWorkers(workersTestList)
		assert.NoError(t, err)

		msg, _ := job.NewMessage("z-0")
		_, err = broker1.PushJob(testInstance, &job.JobRequest{
			WorkerType: "test",
			Message:    msg,
		})
		assert.NoError(t, err)

		go func(broker job.Broker, instance *instance.Instance, n int) {
			for i := 0; i < n; i++ {
				msg, _ := job.NewMessage("a-" + strconv.Itoa(i+1))
				_, err2 := broker.PushJob(instance, &job.JobRequest{
					WorkerType: "test",
					Message:    msg,
				})
				assert.NoError(t, err2)
				time.Sleep(randomMicro(0, v))
			}
		}(broker1, testInstance, n)

		go func(broker job.Broker, instance *instance.Instance, n int) {
			for i := 0; i < n; i++ {
				msg, _ := job.NewMessage("b-" + strconv.Itoa(i+1))
				_, err2 := broker.PushJob(instance, &job.JobRequest{
					WorkerType: "test",
					Message:    msg,
					Manual:     true,
				})
				assert.NoError(t, err2)
				time.Sleep(randomMicro(0, v))
			}
		}(broker2, testInstance, n)

		w.Wait()

		err = broker1.ShutdownWorkers(context.Background())
		assert.NoError(t, err)
		err = broker2.ShutdownWorkers(context.Background())
		assert.NoError(t, err)
		time.Sleep(1 * time.Second)
	})

	t.Run("RedisAddJobRateLimitExceeded", func(t *testing.T) {
		opts1, _ := redis.ParseURL(redisURL1)
		client1 := redis.NewClient(opts1)
		workersTestList := job.WorkersList{
			{
				WorkerType:  "thumbnail",
				Concurrency: 4,
				WorkerFunc: func(ctx *job.WorkerContext) error {
					return nil
				},
			},
		}
		ct := limits.JobThumbnailType
		config.GetRateLimiter().ResetCounter(testInstance, ct)

		broker := job.NewRedisBroker(client1)
		err := broker.StartWorkers(workersTestList)
		assert.NoError(t, err)

		msg, _ := job.NewMessage("z-0")
		j, err := broker.PushJob(testInstance, &job.JobRequest{
			WorkerType: "thumbnail",
			Message:    msg,
		})

		assert.NoError(t, err)
		assert.NotNil(t, j)

		limits.SetMaximumLimit(ct, 10)
		maxLimit := limits.GetMaximumLimit(ct)

		// Blocking the job push
		for i := int64(0); i < maxLimit-1; i++ {
			j, err := broker.PushJob(testInstance, &job.JobRequest{
				WorkerType: "thumbnail",
				Message:    msg,
			})
			assert.NoError(t, err)
			assert.NotNil(t, j)
		}

		j, err = broker.PushJob(testInstance, &job.JobRequest{
			WorkerType: "thumbnail",
			Message:    msg,
		})
		assert.Error(t, err)
		assert.Nil(t, j)
		assert.ErrorIs(t, err, limits.ErrRateLimitReached)

		j, err = broker.PushJob(testInstance, &job.JobRequest{
			WorkerType: "thumbnail",
			Message:    msg,
		})
		assert.Error(t, err)
		assert.Nil(t, j)
	})
}

func randomMicro(min, max int) time.Duration {
	return time.Duration(rand.Intn(max-min)+min) * time.Microsecond
}
