package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	colorTerm "github.com/fatih/color"
)

var (
	Cyan   = colorTerm.New(colorTerm.FgCyan)
	Red    = colorTerm.New(colorTerm.FgRed)
	Yellow = colorTerm.New(colorTerm.FgYellow)
	Green  = colorTerm.New(colorTerm.FgGreen)
)
var colors = map[string]colorTerm.Color{
	"cyan":   *Cyan,
	"red":    *Red,
	"yellow": *Yellow,
	"green":  *Green,
}

type StringSlice []string

func (ss *StringSlice) String() string {
	return strings.Join(*ss, " ")
}
func (ss *StringSlice) Set(val string) error {
	if val == "" {
		return errors.New("no extensions specified")
	}
	stringsSlice := strings.Split(val, ",")

	*ss = append(*ss, stringsSlice...)
	return nil
}

type Response struct {
	url  string
	code int
	err  error
}

func (r *Response) Write() string {
	if r.err != nil {
		return r.url + " :: " + strconv.FormatInt(int64(r.code), 10) + "\n" + "::" + r.err.Error()
	}
	return r.url + " :: " + strconv.FormatInt(int64(r.code), 10) + "\n"
}

type CommonWriter struct {
	responses chan Response
	w         io.Writer
}

func (r *CommonWriter) WriteWithColor(color string) (int, error) {

}
func (r *CommonWriter) Write(w []byte) (int, error) {
	size := 0
	for response := range r.responses {
		data := []byte(response.Write())
		r.w.Write(data)
		size += len(data)
	}
	return size, nil
}

func welcomeDataPrint(method string, gorutines int, target string, extensions []string) {
	Red.Println("_________     _____________       ______\n__  ____/________  __ \\__(_)_________  /_\n_  / __ _  __ \\_  / / /_  /__  ___/_  __ \\\n/ /_/ / / /_/ /  /_/ /_  / _  /   _  /_/ /\n\\____/  \\____//_____/ /_/  /_/    /_.___/")
	fmt.Println()
	Cyan.Print("HTTP Method: ")
	Green.Print(method)
	Yellow.Print(" | ")
	Cyan.Print("Gorutines: ")
	Green.Print(gorutines)
	Yellow.Print(" | ")
	Cyan.Print("Extensions: ")
	Green.Println(strings.Join(extensions, " "))
	fmt.Println()
	Cyan.Print("Target: ")
	Green.Println(target)
	fmt.Println()
	Cyan.Println(":::Starting:::")
	fmt.Println("+---------------+")

}
func endDataPrint(wordsize int64, donesize int64, elapsedTime time.Duration) {
	fmt.Println("+---------------+")
	fmt.Println(":::Completed:::")
	fmt.Println("Recieved codes from :", donesize, "out of:", wordsize, "searches")
	fmt.Println("Elapsed time:", elapsedTime)
}

func removeCharacters(input string, characters string) string {
	filter := func(r rune) rune {
		if strings.IndexRune(characters, r) < 0 {
			return r
		}
		return -1
	}
	return strings.Map(filter, input)
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

//Optimize
func sendRequest(wg *sync.WaitGroup, url string, keyword string, extensions []string, data chan Response) error {
	if strings.Contains(keyword, "%EXT%") {
		keyword = removeCharacters(keyword, "%EXT%")
		for _, ext := range extensions {
			resp, err := http.Get(url + keyword + "." + ext)
			if err != nil && strings.Contains(err.Error(), "connection refused") {
				wg.Done()
				return err
			}
			data <- Response{url: resp.Request.URL.String(), code: resp.StatusCode, err: err}

		}
		wg.Done()
		return nil
	}
	resp, err := http.Get(url + keyword)
	if err != nil && strings.Contains(err.Error(), "connection refused") {
		wg.Done()
		return err
	}
	data <- Response{url: resp.Request.URL.String(), code: resp.StatusCode, err: err}
	wg.Done()
	return nil
}

func sendRequestWorker(wg *sync.WaitGroup, url string, keywords chan string, extensions []string, data chan Response, size *int64) {
	var wgLocal sync.WaitGroup
	for keyword := range keywords {
		wgLocal.Add(1)
		err := sendRequest(&wgLocal, url, keyword, extensions, data)
		if err != nil && strings.Contains(err.Error(), "connection refused") {
			fmt.Println(url+keyword, "::", "connection refused")
			wg.Done()
			return
		}
		atomic.AddInt64(size, 1)
	}
	wg.Done()
}

func bruteWebSite(url string, dict string, extensions []string, power int, visual bool) bool {
	responses := make(chan Response, 5)
	keywords := make(chan string, 50)
	var size int64
	var tSize int64
	var tSizeP *int64
	var sizeP *int64
	sizeP = &size
	tSizeP = &tSize

	power *= 10
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
		go sendRequestWorker(&wg, url, keywords, extensions, responses, tSizeP)
	}
	go func() {
		wg.Wait()
		close(responses)
	}()
	welcomeDataPrint("get", power, url, extensions)
	cw := CommonWriter{responses: responses, w: os.Stdout}
	fmt.Fprint(&cw)
	elapsedTime := time.Since(timer)
	endDataPrint(size, tSize, elapsedTime)
	return true
}
