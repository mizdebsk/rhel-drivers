package cache

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestCacheGet(t *testing.T) {
	t.Parallel()

	type testCase[T any] struct {
		name           string
		cache          func() *Cache[T]
		compute        func() (T, error)
		expectedVal    T
		expectedErr    error
		concurrentRuns int
	}

	errCompute := errors.New("compute failure")
	tests := []testCase[int]{
		{
			name: "returns computed value when cache is empty",
			cache: func() *Cache[int] {
				return &Cache[int]{}
			},
			compute: func() (int, error) {
				return 42, nil
			},
			expectedVal: 42,
			expectedErr: nil,
		},
		{
			name: "returns cached value when cache is ready",
			cache: func() *Cache[int] {
				return &Cache[int]{
					ready: true,
					val:   99,
				}
			},
			compute: func() (int, error) {
				return 0, nil
			},
			expectedVal: 99,
			expectedErr: nil,
		},
		{
			name: "returns error when compute fails",
			cache: func() *Cache[int] {
				return &Cache[int]{}
			},
			compute: func() (int, error) {
				return 0, errCompute
			},
			expectedVal: 0,
			expectedErr: errCompute,
		},
		{
			name: "ensures thread-safe writes to cache under concurrent access",
			cache: func() *Cache[int] {
				return &Cache[int]{}
			},
			compute: func() (int, error) {
				time.Sleep(10 * time.Millisecond)
				return 88, nil
			},
			expectedVal:    88,
			expectedErr:    nil,
			concurrentRuns: 100,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cache := tc.cache()
			var mu sync.Mutex
			var actualResults []struct {
				val int
				err error
			}

			if tc.concurrentRuns > 0 {
				wg := sync.WaitGroup{}
				for i := 0; i < tc.concurrentRuns; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						val, err := cache.Get(tc.compute)
						mu.Lock()
						actualResults = append(actualResults, struct {
							val int
							err error
						}{val, err})
						mu.Unlock()
					}()
				}
				wg.Wait()

				for _, result := range actualResults {
					if result.val != tc.expectedVal || !errors.Is(result.err, tc.expectedErr) {
						t.Errorf("unexpected result in concurrent run: got val=%v, err=%v; want val=%v, err=%v",
							result.val, result.err, tc.expectedVal, tc.expectedErr)
					}
				}
			} else {
				val, err := cache.Get(tc.compute)
				if val != tc.expectedVal || !errors.Is(err, tc.expectedErr) {
					t.Errorf("unexpected result: got val=%v, err=%v; want val=%v, err=%v",
						val, err, tc.expectedVal, tc.expectedErr)
				}
			}
		})
	}
}
