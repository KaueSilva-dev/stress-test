package loadtest


import (
    "context"
    "errors"
	"io"
    "net"
    "net/http"
    "net/url"
    "sort"
    "sync"
    "sync/atomic"
    "time"
)

type Config struct {
	URL string
	Total int
	Concurrency int
	Timeout time.Duration
	Method string
}

type Summary struct {
	TotalDone int
	Status200 int
	Errors int
	StatusMap map[int]int
	StartedAt time.Time
	EndedAt time.Time
}

type StatusCount struct {
	Code int
	Count int
}

func (s Summary) StatusDitributionSorted() []StatusCount {
	arr := make([]StatusCount, 0, len(s.StatusMap))
	for code, count := range s.StatusMap {
		arr = append(arr, StatusCount{Code: code, Count: count})
	}
	sort.Slice(arr, func(i,j int) bool {
		return arr[i].Code < arr[j].Code
	})
	return arr
}

type Event interface{}

type Started struct {}
type RequestDone struct {
	Status int
	Latency time.Duration
	Err error
}
type Finished struct {
	Summary Summary
}

func Run (ctx context.Context, cfg Config) (Summary, error) {
	events := StartAsync(ctx, cfg)
	var sum Summary
	var haveSummary bool

	for ev := range events {
		switch e := ev.(type) {
		case Started:
		
		case RequestDone:
			_ = e
		case Finished:
			sum = e.Summary
			haveSummary = true
		}
	}
	if !haveSummary {
		return sum, errors.New("execução não completou")
	}
	return sum, nil
}

func StartAsync (ctx context.Context, cfg Config) <- chan Event {
	out := make(chan Event, 1024)

	go func() {
		defer close(out)
		out <- Started{}

		parsed, err := url.Parse(cfg.URL)
		if err != nil {
			out <- Finished{Summary: Summary{
				StatusMap: map[int]int{},
			}}
			return
		}

		if cfg.Method == "" {
			cfg.Method = "GET"
		}
		if cfg.Timeout <= 0 {
            cfg.Timeout = 15 * time.Second
        }
        if cfg.Concurrency <= 0 {
            cfg.Concurrency = 1
        }
        if cfg.Total <= 0 {
            cfg.Total = 1
        }

        transport := &http.Transport{
            Proxy: http.ProxyFromEnvironment,
            DialContext: (&net.Dialer{
                Timeout:   5 * time.Second,
                KeepAlive: 30 * time.Second,
            }).DialContext,
            MaxIdleConns:          cfg.Concurrency * 2,
            MaxIdleConnsPerHost:   cfg.Concurrency * 2,
            IdleConnTimeout:       90 * time.Second,
            TLSHandshakeTimeout:   5 * time.Second,
            ExpectContinueTimeout: 1 * time.Second,
        }

        client := &http.Client{
            Transport: transport,
            Timeout:   cfg.Timeout,
        }

        // Canal de jobs
        jobs := make(chan int, cfg.Total)
        go func() {
            for i := 0; i < cfg.Total; i++ {
                jobs <- i
            }
            close(jobs)
        }()

        var totalDone int64
        var status200 int64
        var errorsCount int64

        statusMap := sync.Map{}

        start := time.Now()

        wg := sync.WaitGroup{}
        wg.Add(cfg.Concurrency)

        for w := 0; w < cfg.Concurrency; w++ {
            go func() {
                defer wg.Done()
                for range jobs {
                    select {
                    case <-ctx.Done():
                        return
                    default:
                    }

                    req, err := http.NewRequest(cfg.Method, parsed.String(), nil)
                    if err != nil {
                        atomic.AddInt64(&errorsCount, 1)
                        out <- RequestDone{Status: 0, Latency: 0, Err: err}
                        continue
                    }

                    t0 := time.Now()
                    resp, err := client.Do(req)
                    lat := time.Since(t0)

                    if err != nil {
                        atomic.AddInt64(&errorsCount, 1)
                        out <- RequestDone{Status: 0, Latency: lat, Err: err}
                        continue
                    }

                    _ = drainAndClose(resp.Body)

                    atomic.AddInt64(&totalDone, 1)
                    if resp.StatusCode == http.StatusOK {
                        atomic.AddInt64(&status200, 1)
                    }

                    incrementStatus(&statusMap, resp.StatusCode)
                    out <- RequestDone{Status: resp.StatusCode, Latency: lat, Err: nil}
                }
            }()
        }

        wg.Wait()
        end := time.Now()

        summary := Summary{
            TotalDone: int(atomic.LoadInt64(&totalDone)),
            Status200: int(atomic.LoadInt64(&status200)),
            Errors:    int(atomic.LoadInt64(&errorsCount)),
            StatusMap: mapFromSyncMap(&statusMap),
            StartedAt: start,
            EndedAt:   end,
        }

        out <- Finished{Summary: summary}
    }()

    return out
}

func drainAndClose(body io.ReadCloser) error {
    if body == nil {
        return nil
    }
    _, _ = io.Copy(io.Discard, body)
    return body.Close()
}

type netConnCloser interface {
    Close() error
}

func incrementStatus(m *sync.Map, code int) {
    actual, _ := m.LoadOrStore(code, new(int64))
    p := actual.(*int64)
    atomic.AddInt64(p, 1)
}

func mapFromSyncMap(m *sync.Map) map[int]int {
    out := make(map[int]int)
    m.Range(func(k, v any) bool {
        code := k.(int)
        count := int(atomic.LoadInt64(v.(*int64)))
        out[code] = count
        return true
    })
    return out
}