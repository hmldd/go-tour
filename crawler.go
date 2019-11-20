/*
	A Tour of Go Exercise: Web Crawler
	Go语言之旅 - 网络爬虫
	https://tour.golang.org/concurrency/10
*/
package main

import (
	"fmt"
	"sync"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

// 使用sync.Mutex对数据加锁，使用sync.WaitGroup等待所有goroutine执行完毕
type UrlCounter struct {
	v   map[string]int
	mux sync.Mutex
	wg  sync.WaitGroup
}

func (u *UrlCounter) isVisited(url string) bool {
	u.mux.Lock()
	defer u.mux.Unlock()
	u.v[url]++
	// 如果count大于1说明已经抓取过，返回true
	if count := u.v[url]; count > 1 {
		return true
	}

	// 未抓取过，返回false
	return false
}

// 全局变量，多个goroutine使用时需要加锁
// 实际项目中应避免使用全局变量
// var uc = &UrlCounter{v: make(map[string]int)}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, uc *UrlCounter) {
	// 抓取结束后通知WaitGroup
	defer uc.wg.Done()
	if uc.isVisited(url) {
		return
	}

	if depth <= 0 {
		return
	}
	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("found: %s %q\n", url, body)
	for _, u := range urls {
		// 启动goroutine时WaitGroup加1
		uc.wg.Add(1)
		go Crawl(u, depth-1, fetcher, uc)
	}

	return
}

func main() {
	uc := &UrlCounter{v: make(map[string]int)}
	// 所有Crawl方法都调用了wg.Done()，所以这里需要加1
	uc.wg.Add(1)
	Crawl("https://golang.org/", 4, fetcher, uc)
	// 等待所有goroutine执行完毕
	uc.wg.Wait()
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}
