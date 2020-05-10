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

func welcomeDataPrint(method string, gorutines int, target string) {
	fmt.Println(" _|. _ _  _  _  _ _|_\n(_||| _) (/_(_|| (_| )\n[logo officually stolen from original dirbsearch]")
	fmt.Println("Method:", method, "|", "Gorutines:", gorutines)
	fmt.Println("Target:", target, "\n")
	fmt.Println("Starting")
}
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

// Atomic - poor implementation -> create size, size_p ??!?
func scanDict(filename string, keywords chan string, size *int64) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		keywords <- scanner.Text()
		*size++
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
func bruteWebSite(url string, dict string, visual bool) bool {
	responses := make(chan Response, 5)
	keywords := make(chan string, 50)
	// Is needed to print size of word_count
	// Не сработает, если много слов, потому что уже начнётся sendRequest() и, если не начать выводить
	// 	инфу сразу по получении данных из канала responses, то первые ответы начнут теряться, поэтому нельзя ждать
	// 	до конца
	// НЕЛЬЗЯ ВЫВЕСТИ SIZE в начале, т.к неизвестно кол-во слов
	var size int64
	var sizeP *int64
	size = 0
	sizeP = &size

	go scanDict(dict, keywords, sizeP)

	var wg sync.WaitGroup
	for keyword := range keywords {
		wg.Add(1)
		go sendRequest(&wg, url, keyword, responses)
	}
	go func() {
		wg.Wait()
		close(responses)
	}()
	// Кол-во горутин зависит от мощности поиска, который задаётся флагом, по дефолту 10-20 горутин
	welcomeDataPrint("get", 10, url)
	for response := range responses {
		response.printResponse(1)
	}
	// for response := range responses {
	// 	response.printResponse(1)
	// }

	return true
}
