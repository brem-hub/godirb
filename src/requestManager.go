package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type requestManager struct {
	logger     *loggerCust
	w          io.Writer
	url        string
	method     string
	keywords   chan string
	dict       string
	power      int
	depth      int
	timer      int
	responses  chan Response
	sizeP      *int64
	tSizeP     *int64
	interrupt  chan os.Signal
	extensions []string
}

func (rm *requestManager) New(w io.Writer, logger *loggerCust, url string, method string, keywords chan string, dict string, power int, depth int, timer int, responses chan Response, extensions []string, sizeP *int64, tSizeP *int64) {
	rm.logger = logger
	rm.w = w
	rm.url = url
	rm.method = method
	rm.keywords = keywords
	rm.dict = dict
	rm.power = power
	rm.depth = depth
	rm.timer = timer
	rm.responses = responses
	rm.sizeP = sizeP
	rm.tSizeP = tSizeP
	rm.extensions = extensions
}
func (rm *requestManager) Start(ctx context.Context, cancel context.CancelFunc) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	rm.workerLauncher(ctx, cancel, interrupt)
}

func (rm *requestManager) scanDict(ctx context.Context, recursive bool) ([]string, chan error) {
	errc := make(chan error, 1)
	file, err := os.Open(rm.dict)
	cache := make([]string, 1)
	errc <- err
	if err != nil {
		return []string{}, errc
	}
	scanner := bufio.NewScanner(file)
	defer close(rm.keywords)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return []string{}, errc
		default:
			text := scanner.Text()
			if recursive {
				cache = append(cache, text)
			}
			rm.keywords <- text
			*rm.sizeP++
		}
	}
	return cache, errc
}

func sliceToChan(slice []string, ch chan string) {
	for _, el := range slice {
		ch <- el
	}
	close(ch)
}

func (rm *requestManager) workerLauncher(ctx context.Context, cancel context.CancelFunc, interrupt chan os.Signal) {
	var wg sync.WaitGroup
	var wgT sync.WaitGroup
	var vCache []string
	depth := rm.depth
	vTime := time.Duration(rm.timer) * time.Second
	recursiveFlag := false
	recursiveChan := make(chan map[string]int, 10)

	if rm.depth > 1 {
		recursiveFlag = true
	}
	go func() {
		for range interrupt {
			cancel()
			// To remove C^
			fmt.Print("\r \r")
			if rm.w != os.Stdout {
				fmt.Fprint(rm.w, ":::Canceled by user:::")
			}
			RedWhite.Fprint(os.Stdout, "Canceled by user")
			fmt.Println()
			rm.logger.Println("Canceled by user")
			os.Exit(1)
		}

	}()
	wgT.Add(1)
	go func() {
		cache, errc := rm.scanDict(ctx, recursiveFlag)
		err := <-errc
		if err != nil {
			rm.logger.logger.Fatalln(err)
		}
		vCache = cache
		wgT.Done()
	}()
	// wgT.Wait()
	for grNum := 0; grNum < rm.power; grNum++ {
		wg.Add(1)
		if vTime > 0 {
			go rm.requestWorkerWithTimer(ctx, &wg, rm.url, rm.keywords, vTime, recursiveChan, depth)
		} else {
			go rm.requestWorker(ctx, &wg, rm.url, rm.keywords, recursiveChan, depth)
		}
	}
	go func() {
		for recursive := range recursiveChan {
			ch := make(chan string, 100)
			go sliceToChan(vCache, ch)
			for urlI, depth := range recursive {
				wgT.Add(1)
				if vTime > 0 {
					go rm.requestWorkerWithTimer(ctx, &wgT, urlI, ch, vTime, recursiveChan, depth)
				} else {
					go rm.requestWorker(ctx, &wgT, urlI, ch, recursiveChan, depth)
				}
			}
		}
	}()
	go func() {
		wg.Wait()
		wgT.Wait()
		close(rm.responses)
		close(recursiveChan)
	}()
}

func (rm *requestManager) requestWorkerWithTimer(ctx context.Context, wg *sync.WaitGroup, url string, keywords chan string, rTime time.Duration, recursive chan map[string]int, depth int) {
	timer := time.NewTimer(rTime)
	for keyword := range keywords {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		case <-timer.C:
			urls, err := sendRequest(url, rm.method, rm.logger, keyword, rm.extensions, rm.responses)
			if err != nil {
				rm.logger.Println(err.Error())
				Red.Println("Error occured")
				wg.Done()
				return
			}
			atomic.AddInt64(rm.tSizeP, 1)

			if len(urls) > 0 && depth > 1 {
				depth--
				for _, urlI := range urls {
					recursive <- map[string]int{urlI: depth}
				}
			}
			timer.Reset(rTime)
		}
	}
	wg.Done()
}

func (rm *requestManager) requestWorker(ctx context.Context, wg *sync.WaitGroup, url string, keywords chan string, recursive chan map[string]int, depth int) {
	for keyword := range keywords {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		default:
			urls, err := sendRequest(url, rm.method, rm.logger, keyword, rm.extensions, rm.responses)
			if err != nil {
				rm.logger.Println(err.Error())
				Red.Println("Error occured")
				wg.Done()
				return
			}
			atomic.AddInt64(rm.tSizeP, 1)

			if len(urls) > 0 && depth > 1 {
				depth--
				for _, urlI := range urls {
					recursive <- map[string]int{urlI: depth}
				}
			}
		}
	}
	wg.Done()
}
func getRequestCustom(url string, keyword string, method string) (Response, error) {
	req, err := http.NewRequest(strings.ToUpper(method), url, nil)
	if err != nil {
		return Response{}, err
	}
	res, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		var errr error
		switch {
		case strings.Contains(err.Error(), "connection refused"):
			errr = errors.New(url + "/" + keyword + " :: connection refused")
		case strings.Contains(err.Error(), "protocol scheme"):
			errr = errors.New(url + "/" + keyword + " :: unsupported protocol scheme")
		case strings.Contains(err.Error(), "dial tcp"):
			errr = errors.New(url + "/" + keyword + " :: no such host (dial tcp)")
		default:
			errr = errors.New(url + "/" + keyword + " :: not custom error :: " + err.Error())
		}
		return Response{}, errr
	}
	defer res.Body.Close()
	size := res.ContentLength
	if size == -1 {
		size = 0
	}
	return Response{keyword: keyword, url: res.Request.URL.String(), code: res.StatusCode, size: size}, nil
}

func sendRequest(url string, method string, logger *loggerCust, keyword string, extensions []string, data chan Response) ([]string, error) {
	codesToRecursive := []int{200, 301, 302, 303, 307, 308}
	if keyword == "" {
		return []string{}, nil
	}
	urls := make([]string, 0)

	if strings.Contains(keyword, "%EXT%") {
		var fullExt string
		for _, ext := range extensions {
			var resp Response
			var err error
			if ext == "none" {
				resp, err = getRequestCustom(url, fullExt, method)
			} else {
				fullExt = strings.Replace(keyword, "%EXT%", "."+ext, 1)
				resp, err = getRequestCustom(url+"/"+fullExt, fullExt, method)
			}
			if err != nil {
				logger.Println(err.Error())
				return []string{}, err
			}
			if checkCodes(resp.code, codesToRecursive) {
				urls = append(urls, resp.url)
			}
			data <- resp
		}
	} else {
		resp, err := getRequestCustom(url+"/"+keyword, keyword, method)
		if err != nil {
			logger.Println(err.Error())
			return []string{}, err
		}
		data <- resp

		if checkCodes(resp.code, codesToRecursive) {
			urls = append(urls, resp.url)
		}
	}
	return urls, nil
}
