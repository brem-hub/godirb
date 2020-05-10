package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Response struct {
	url  string
	code int
	err  error
}

func (r *Response) Write() string {
	return r.url + " :: " + strconv.FormatInt(int64(r.code), 10) + "\n"
}

type CommonWriter struct {
	responses chan Response
	w         io.Writer
}

func (r *CommonWriter) Write(w []byte) (int, error) {
	for response := range r.responses {
		r.w.Write([]byte(response.Write()))
	}
	return 1, nil
}

func welcomeDataPrint(method string, gorutines int, target string) {
	fmt.Println(" _|. _ _  _  _  _ _|_\n(_||| _) (/_(_|| (_| )\n[logo is used from original dirbsearch]")
	fmt.Println("Method:", method, "|", "Gorutines:", gorutines)
	fmt.Println("Target:", target)
	fmt.Println()
	fmt.Println(":::Starting:::")
	fmt.Println("+---------------+")

}
func endDataPrint(wordsize int64, donesize int64, elapsedTime time.Duration) {
	fmt.Println("+---------------+")
	fmt.Println(":::Completed:::")
	fmt.Println("Recieved codes from :", donesize, "out of:", wordsize, "searches")
	fmt.Println("Elapsed time:", elapsedTime)
}
func scanDict(filename string, keywords chan string, size *int64) chan error {
	errc := make(chan error, 1)
	file, err := os.Open(filename)
	errc <- err
	if err != nil {
		return errc
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		keywords <- scanner.Text()
		*size++
	}
	close(keywords)
	return errc
}
func errorPrintAndExit(err error) {
	fmt.Println(err)
	os.Exit(1)
}
func errorPrint(err error) {
	fmt.Println(err)
}
func sendRequest(wg *sync.WaitGroup, url string, keyword string, data chan Response) error {
	resp, err := http.Get(url + keyword)
	if err != nil && strings.Contains(err.Error(), "connection refused") {
		wg.Done()
		return err
	}
	data <- Response{url: resp.Request.URL.String(), code: resp.StatusCode, err: err}
	wg.Done()
	return err
}

func sendRequestWorker(wg *sync.WaitGroup, url string, keywords chan string, data chan Response, size *int64) {
	var wgLocal sync.WaitGroup
	for keyword := range keywords {
		wgLocal.Add(1)
		err := sendRequest(&wgLocal, url, keyword, data)
		if err != nil && strings.Contains(err.Error(), "connection refused") {
			fmt.Println(url+keyword, "::", "connection refused")
			wg.Done()
			return
		}
		atomic.AddInt64(size, 1)
	}
	wg.Done()
}

func bruteWebSite(url string, dict string, power int, visual bool) bool {
	responses := make(chan Response, 5)
	keywords := make(chan string, 50)
	/*
		Не сработает, если много слов, потому что уже начнётся sendRequest() и, если не начать выводить
			инфу сразу по получении данных из канала responses, то первые ответы начнут теряться, поэтому нельзя ждать
			до конца
	*/
	var size int64
	var tSize int64
	var tSizeP *int64
	var sizeP *int64
	size = 0
	sizeP = &size
	tSizeP = &tSize

	power = power * 10
	timer := time.Now()

	go func() {
		errc := scanDict(dict, keywords, sizeP)
		err := <-errc
		if err != nil {
			errorPrintAndExit(err)
		}
	}()

	var wg sync.WaitGroup
	for grNum := 0; grNum < power; grNum++ {
		wg.Add(1)
		go sendRequestWorker(&wg, url, keywords, responses, tSizeP)
	}
	go func() {
		wg.Wait()
		close(responses)
	}()
	welcomeDataPrint("get", power, url)
	cw := CommonWriter{responses: responses, w: os.Stdout}
	fmt.Fprint(&cw)
	elapsedTime := time.Since(timer)
	endDataPrint(size, tSize, elapsedTime)
	return true
}
