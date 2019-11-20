// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
	Go Codewalk: Share Memory By Communicating
	Go 代码漫读 - 通过通信共享内存
	https://golang.org/doc/codewalk/sharemem/
*/

package main

import (
	"log"
	"net/http"
	"time"
)

const (
	numPollers     = 2                // number of Poller goroutines to launch 同时拉取的goroutine数量
	pollInterval   = 60 * time.Second // how often to poll each URL 每隔过久拉取一遍URL
	statusInterval = 10 * time.Second // how often to log status to stdout 每隔多久打印一次状态至stdout
	errTimeout     = 10 * time.Second // back-off timeout on error 报错后推迟多久
)

var urls = []string{
	"http://www.google.com/",
	"http://golang.org/",
	"http://blog.golang.org/",
}

// State represents the last-known state of a URL.
// State 表示URL最后一次拉取时的状态。
type State struct {
	url    string
	status string
}

// StateMonitor maintains a map that stores the state of the URLs being
// polled, and prints the current state every updateInterval nanoseconds.
// It returns a chan State to which resource state should be sent.
// StateMonitor 维护一个map，记录URL拉取的状态，每隔updateInterval纳秒打印一次当前状态。
// 它返回一个State类型的channel给需要更新状态的资源。
func StateMonitor(updateInterval time.Duration) chan<- State {
	updates := make(chan State)
	urlStatus := make(map[string]string)
	ticker := time.NewTicker(updateInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				logState(urlStatus)
			case s := <-updates:
				urlStatus[s.url] = s.status
			}
		}
	}()
	return updates
}

// logState prints a state map.
// logState 打印记录状态的map。
func logState(s map[string]string) {
	log.Println("Current state:")
	for k, v := range s {
		log.Printf(" %s %s", k, v)
	}
}

// Resource represents an HTTP URL to be polled by this program.
// Resource 表示一个需要被拉取的HTTP URL资源。
type Resource struct {
	url      string
	errCount int
}

// Poll executes an HTTP HEAD request for url
// and returns the HTTP status string or an error string.
// Poll 为url执行一个HTTP HEAD请求
// 并返回表示HTTP状态或错误的字符串。
func (r *Resource) Poll() string {
	resp, err := http.Head(r.url)
	if err != nil {
		log.Println("Error", r.url, err)
		r.errCount++
		return err.Error()
	}
	r.errCount = 0
	return resp.Status
}

// Sleep sleeps for an appropriate interval (dependent on error state)
// before sending the Resource to done.
// Sleep 休眠适当的时间（根据报错的状态）
// 然后再发送资源给done队列
func (r *Resource) Sleep(done chan<- *Resource) {
	time.Sleep(pollInterval + errTimeout*time.Duration(r.errCount))
	done <- r
}

func Poller(in <-chan *Resource, out chan<- *Resource, status chan<- State) {
	for r := range in {
		s := r.Poll()
		status <- State{r.url, s}
		out <- r
	}
}

func main() {
	// Create our input and output channels.
	// 创建pending和complete队列。
	pending, complete := make(chan *Resource), make(chan *Resource)

	// Launch the StateMonitor.
	// 启动StateMonitor监控状态。
	status := StateMonitor(statusInterval)

	// Launch some Poller goroutines.
	// 启动多个Poller goroutine。
	for i := 0; i < numPollers; i++ {
		go Poller(pending, complete, status)
	}

	// Send some Resources to the pending queue.
	// 发送一些Resource给pending队列。
	go func() {
		for _, url := range urls {
			pending <- &Resource{url: url}
		}
	}()

	// 已经完成的Resource延迟适当的时间后发送给pending队列循环更新。
	for r := range complete {
		go r.Sleep(pending)
	}
}
