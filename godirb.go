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

	colorTerm "github.com/fatih/color"
)

var (
	Cyan     = colorTerm.New(colorTerm.FgCyan)
	Blue     = colorTerm.New(colorTerm.FgBlue)
	Red      = colorTerm.New(colorTerm.FgRed)
	Yellow   = colorTerm.New(colorTerm.FgYellow)
	Green    = colorTerm.New(colorTerm.FgGreen)
	GreenBg  = colorTerm.New(colorTerm.FgYellow, colorTerm.BgGreen)
	RedWhite = colorTerm.New(colorTerm.FgWhite, colorTerm.BgRed)
)

var colors = map[int]*colorTerm.Color{
	404: Cyan,
	200: GreenBg,
	301: GreenBg,
	302: GreenBg,
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
	url     string
	code    int
	size    int64
	keyword string
}

func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%3d   B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func (r *Response) Write() string {
	size := ByteCountIEC(r.size)
	str := fmt.Sprintf("%s  ::  %3d  ::  /%s/  ->  %s", size, r.code, r.keyword, r.url)
	return str
}

type CommonWriter struct {
	responses chan Response
	w         io.Writer
}

func (r *CommonWriter) writeWithColors(ctx context.Context, verbose bool) (int, error) {
	size := 0
	tmp := 0
	for response := range r.responses {
		select {
		case <-ctx.Done():
			fmt.Fprintln(r.w)
			return size, nil
		default:
			var color *colorTerm.Color
			if _, ok := colors[response.code]; !ok {
				color = Cyan
			} else {
				color = colors[response.code]
			}
			if verbose {
				color.Fprint(r.w, response.Write())
				fmt.Println()
			} else {
				if response.code != 404 {
					color.Fprint(r.w, response.Write())
					fmt.Println()
					size += tmp
				}
			}
		}

	}
	return size, nil
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

func welcomeDataPrint(w io.Writer, method string, gorutines int, target string, extensions []string) {
	Blue.Fprintln(w, "_________     _____________       ______\n__  ____/________  __ \\__(_)_________  /_\n_  / __ _  __ \\_  / / /_  /__  ___/_  __ \\\n/ /_/ / / /_/ /  /_/ /_  / _  /   _  /_/ /\n\\____/  \\____//_____/ /_/  /_/    /_.___/")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s %s %s %s %s %s %s %s\n\n", Blue.Sprint("HTTP Method:"), Green.Sprint(method), Yellow.Sprint("|"),
		Blue.Sprint("Gorutines:"), Green.Sprint(gorutines), Yellow.Sprint("|"),
		Blue.Sprint("Extensions:"), Green.Sprint(strings.Join(extensions, " ")))
	fmt.Fprintf(w, "%s %s\n\n", Blue.Sprint("Target:"), Green.Sprint(target))
	Blue.Fprintln(w, ":::Starting:::")
	fmt.Fprintln(w, "+---------------+")
}
func endDataPrint(w io.Writer, wordsize int64, donesize int64, elapsedTime time.Duration) {
	fmt.Fprintln(w, "+---------------+")
	Blue.Fprintln(w, ":::Completed:::")
	fmt.Fprintln(w, "Recieved codes from :", donesize, "out of:", wordsize, "searches")
	fmt.Fprintln(w, "Elapsed time:", elapsedTime)
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
func scanDict(ctx context.Context, filename string, keywords chan string, size *int64) chan error {
	errc := make(chan error, 1)
	file, err := os.Open(filename)
	errc <- err
	if err != nil {
		return errc
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			close(keywords)
			return errc
		default:
			keywords <- scanner.Text()
			*size++
		}
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

func getRequestCustom(url string, keyword string) (Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Response{}, err
	}
	res, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return Response{}, err
	}
	return Response{keyword: keyword, url: res.Request.URL.String(), code: res.StatusCode, size: res.ContentLength}, nil
}

func sendRequest(url string, keyword string, extensions []string, data chan Response) error {
	if strings.Contains(keyword, "%EXT%") {
		keyword = removeCharacters(keyword, "%EXT%")
		for _, ext := range extensions {
			resp, err := getRequestCustom(url+keyword+"."+ext, keyword+"."+ext)
			if err != nil && strings.Contains(err.Error(), "connection refused") {
				return errors.New(resp.url + " :: connection refused")
			}
			data <- resp
		}
		return nil
	}
	resp, err := getRequestCustom(url+keyword, keyword)
	if err != nil && strings.Contains(err.Error(), "connection refused") {
		return errors.New(url + keyword + " :: connection refused")

	}
	data <- resp

	return nil
}

//??? Why it doesn`t stop when connection refused is sent
func requestWorker(ctx context.Context, wg *sync.WaitGroup, url string, keywords chan string, extensions []string, data chan Response, size *int64) {
	// flag := false
	// for {
	// 	errc := make(chan error, 1)
	// 	select {
	// 	case keyword, ok := <-keywords:
	// 		go sendRequest(url, keyword, extensions, data)
	// 		if !ok {
	// 			fmt.Println("EDEDEEEEEDDDE")
	// 			keywords = nil
	// 		}
	// 	case err := <-errc:
	// 		fmt.Println(err)
	// 		wg.Done()
	// 		return
	// 	}
	// 	if flag {
	// 		wg.Done()
	// 		return
	// 	}

	for keyword := range keywords {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		default:
			err := sendRequest(url, keyword, extensions, data)
			if err != nil {
				fmt.Println(err)
				break
			}
			atomic.AddInt64(size, 1)
		}
	}
	wg.Done()
}

func workerLauncher(ctx context.Context, cancel context.CancelFunc, url string, keywords chan string, dict string, power int, responses chan Response, sizeP *int64, tSizeP *int64, interrupt chan os.Signal) {
	var wg sync.WaitGroup

	go func() {
		errc := scanDict(ctx, dict, keywords, sizeP)
		err := <-errc
		if err != nil {
			errorPrintAndExit(err)
		}
	}()

	for grNum := 0; grNum < power; grNum++ {
		wg.Add(1)
		go requestWorker(ctx, &wg, url, keywords, extensions, responses, tSizeP)
	}
	go func() {
		for range interrupt {
			cancel()
			// To remove C^
			fmt.Print("\r \r")
			RedWhite.Print("Canceled by user")
		}

	}()
	go func() {
		wg.Wait()
		close(responses)
	}()
}
func bruteWebSite(url string, dict string, extensions []string, method string, power int, verbose bool, w io.Writer) bool {
	responses := make(chan Response, 5)
	keywords := make(chan string, 50)
	var size int64
	var tSize int64
	var tSizeP *int64
	var sizeP *int64
	sizeP = &size
	tSizeP = &tSize

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	power *= 10
	timer := time.Now()
	cw := CommonWriter{responses: responses, w: w}

	workerLauncher(ctx, cancel, url, keywords, dict, power, responses, sizeP, tSizeP, interrupt)
	welcomeDataPrint(w, method, power, url, extensions)
	cw.writeWithColors(ctx, verbose)
	endDataPrint(w, size, tSize, time.Since(timer))

	return true
}
