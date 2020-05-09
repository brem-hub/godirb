package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"sync"
)

type Response struct {
	url  string
	code int
}

//Deep
func (r Response) printResponse(deep int) {
	switch deep {
	case 1:
		fmt.Println(r.url, ": ", r.code)
	case 2:
		if r.code == 404 {
			fmt.Println(r.url, ": ", r.code)
		}
	}
}
func scanDict(filename string, keywords chan string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		keywords <- scanner.Text()
	}
	close(keywords)
}

func sendRequest(wg *sync.WaitGroup, url string, keyword string, data chan Response) {
	resp, err := http.Get(url + keyword)
	if err != nil {
		fmt.Println(err)
	}
	data <- Response{url: resp.Request.URL.String(), code: resp.StatusCode}
	wg.Done()
}

//Depth of printing (each file or just 404 ) -> check Response{}
func bruteWebSite(url string, dict string) bool {
	responses := make(chan Response, 100)
	keywords := make(chan string, 50)
	go scanDict(dict, keywords)
	var wg sync.WaitGroup

	for keyword := range keywords {
		wg.Add(1)
		go sendRequest(&wg, url, keyword, responses)
	}
	go func() {
		wg.Wait()
		close(responses)

	}()
	for response := range responses {
		response.printResponse(1)
	}
	fmt.Println("done")
	return true
}
