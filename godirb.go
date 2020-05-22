package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
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
	200: Green,
	301: Green,
	302: Green,
	500: Cyan,
}

func checkColors(w io.Writer) {
	if w != os.Stdout {
		Cyan.DisableColor()
		Blue.DisableColor()
		Red.DisableColor()
		Yellow.DisableColor()
		Green.DisableColor()
		GreenBg.DisableColor()
	}
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

type loggerCust struct {
	mutex  sync.Mutex
	logger *log.Logger
	file   *os.File
}

func (l *loggerCust) Println(message string) {
	l.logger.Printf("[%02d:%02d:%02d] %s\n", time.Now().Hour(), time.Now().Minute(), time.Now().Second(), message)
}

func (l *loggerCust) createLogger(url string) {
	path := clearUrl(url)
	timer := time.Now()
	file, err := os.Create(fmt.Sprintf("log/log_%s_%d-%02d-%02d_%02d-%02d-%02d", path, timer.Year(), timer.Month(),
		timer.Day(), timer.Hour(), timer.Minute(), timer.Second()))
	if err != nil {
		fmt.Println(err)
		return
	}
	l.file = file
	log := log.New(file, "", 0)
	l.logger = log
	l.mutex = sync.Mutex{}
	l.Println("Logger is started")
}

func (l *loggerCust) closeLogger() {
	l.Println("Logger is closed")
	l.file.Close()
}
func clearDir(dir string) error {
	names, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range names {
		os.RemoveAll(path.Join([]string{dir, entry.Name()}...))
	}
	return nil
}
func clearUrl(url string) string {
	var path string
	if strings.Contains(url, "127.0.0.1") {
		path = "local"
	} else {
		if strings.Contains(url, ".com") {
			path = strings.Replace(url, ".com", "", -1)
		} else if strings.Contains(url, ".ru") {
			path = strings.Replace(url, ".ru", "", -1)
		}
		if strings.Contains(path, "https://") {
			path = strings.Replace(path, "https://", "", -1)
		} else if strings.Contains(path, "http://") {
			path = strings.Replace(path, "http://", "", -1)
		}
	}
	return path
}

func addHTTPHTTPSProtocols(url string, protocol string) string {
	switch {
	case protocol == "http":
		return "http://" + url
	case protocol == "https":
		return "https://" + url
	}
	return ""
}

type CommonWriter struct {
	responses chan Response
	codes     []int
	w         io.Writer
}

func (r *CommonWriter) checkCodes(code int) bool {
	for _, v := range r.codes {
		if v == code {
			return true
		}
	}
	return false
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
				fmt.Fprintln(r.w)
			} else {
				if !r.checkCodes(response.code) {
					color.Fprintf(r.w, "\r%s", response.Write())
					fmt.Fprintln(r.w)
					size += tmp
				}
			}
		}
	}
	return size, nil
}

//Set speed of animation in Milliseconds
func loader(speed time.Duration) {
	time.Sleep(150 * time.Millisecond)
	for {
		fmt.Printf("\r\\")
		time.Sleep(speed * time.Millisecond)
		fmt.Printf("\r/")
		time.Sleep(speed * time.Millisecond)
	}
}
func welcomeDataPrint(w io.Writer, logger *loggerCust, method string, gorutines int, target string, extensions []string) {
	if len(extensions) == 0 {
		extensions = append(extensions, "none")
	}
	Blue.Fprintln(w, "_________     _____________       ______\n__  ____/________  __ \\__(_)_________  /_\n_  / __ _  __ \\_  / / /_  /__  ___/_  __ \\\n/ /_/ / / /_/ /  /_/ /_  / _  /   _  /_/ /\n\\____/  \\____//_____/ /_/  /_/    /_.___/")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s %s %s %s %s %s %s %s\n\n", Blue.Sprint("HTTP Method:"), Green.Sprint(method), Yellow.Sprint("|"),
		Blue.Sprint("Gorutines:"), Green.Sprint(gorutines), Yellow.Sprint("|"),
		Blue.Sprint("Extensions:"), Green.Sprint(strings.Join(extensions, " ")))
	fmt.Fprintf(w, "%s %s\n\n", Blue.Sprint("Error log:"), Green.Sprint(logger.file.Name()))
	fmt.Fprintf(w, "%s %s\n\n", Blue.Sprint("Target:"), Green.Sprint(target))
	Blue.Fprintln(w, ":::Starting:::")
	fmt.Fprintln(w, "+---------------+")
}
func endDataPrint(w io.Writer, wordsize int64, donesize int64, elapsedTime time.Duration) {
	fmt.Fprintln(w, "\r+---------------+")
	Blue.Fprintln(w, ":::Completed:::")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s/%s codes returned\n", Green.Sprint(donesize), Blue.Sprint(wordsize))
	fmt.Fprintf(w, "Elapsed time: %s\n", Green.Sprint(elapsedTime))
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
func scanDict(ctx context.Context, filename string, keywords chan string, size *int64, recursive bool) ([]string, chan error) {
	errc := make(chan error, 1)
	file, err := os.Open(filename)
	cache := make([]string, 1)
	errc <- err
	if err != nil {
		return []string{}, errc
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			close(keywords)
			return []string{}, errc
		default:
			text := scanner.Text()
			if recursive {
				cache = append(cache, text)
			}
			keywords <- text
			*size++
		}
	}
	close(keywords)
	return cache, errc
}
func getRequestCustom(url string, keyword string) (Response, error) {
	req, err := http.NewRequest("GET", url, nil)
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

func sendRequest(url string, logger *loggerCust, keyword string, extensions []string, data chan Response) ([]string, error) {
	if strings.Contains(keyword, "%EXT%") {
		urls := make([]string, 0)
		for _, ext := range extensions {
			fullExt := strings.Replace(keyword, "%EXT%", "."+ext, 1)
			resp, err := getRequestCustom(url+"/"+fullExt, fullExt)
			if err != nil {
				logger.Println(err.Error())
				return []string{}, err
			}
			if resp.code == 307 || resp.code == 308 {
				urls = append(urls, resp.url)
			}
			data <- resp

		}
		return urls, nil
	}
	resp, err := getRequestCustom(url+"/"+keyword, keyword)
	if err != nil {
		logger.Println(err.Error())
		return []string{}, err
	}
	data <- resp

	if resp.code == 307 || resp.code == 308 {
		return []string{resp.url}, nil
	}
	return []string{resp.url}, nil
}

func requestWorker(ctx context.Context, logger *loggerCust, wg *sync.WaitGroup, url string, keywords chan string, extensions []string, depth int32, data chan Response, recursive chan map[string]int32, size *int64) {
	fmt.Println("WORKING FOR", url)
	for keyword := range keywords {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		default:

			urls, err := sendRequest(url, logger, keyword, extensions, data)
			if err != nil {
				logger.Println(err.Error())
				Red.Println("Error occured")
				wg.Done()
				return
			}
			atomic.AddInt64(size, 1)

			if len(urls) > 0 && depth > 0 {
				depth--
				for _, urlI := range urls {
					recursive <- map[string]int32{urlI: depth}
				}
			}
		}
	}
	wg.Done()
}

func sliceToChan(slice []string) chan string {
	ch := make(chan string, len(slice))
	for _, el := range slice {
		ch <- el
	}
	close(ch)
	return ch
}

func workerLauncher(ctx context.Context, cancel context.CancelFunc, logger *loggerCust, w io.Writer, url string, keywords chan string, dict string, power int, responses chan Response, sizeP *int64, tSizeP *int64, interrupt chan os.Signal) {
	var wg sync.WaitGroup
	var wgT sync.WaitGroup
	recursiveChan := make(chan map[string]int32, 10)
	var depth int32
	var vCache []string
	recursive := true
	depth = 2
	go func() {
		for range interrupt {
			cancel()
			// To remove C^
			fmt.Print("\r \r")
			if w != os.Stdout {
				fmt.Fprint(w, ":::Canceled by user:::")
			}
			RedWhite.Fprint(os.Stdout, "Canceled by user")
			fmt.Println()
			logger.Println("Canceled by user")
			os.Exit(1)
		}

	}()

	go func() {
		cache, errc := scanDict(ctx, dict, keywords, sizeP, recursive)
		err := <-errc
		if err != nil {
			logger.logger.Fatalln(err)
		}
		vCache = cache
	}()

	for grNum := 0; grNum < power; grNum++ {
		wg.Add(1)
		go requestWorker(ctx, logger, &wg, url, keywords, extensions, depth, responses, recursiveChan, tSizeP)
	}

	go func() {
		for recursive := range recursiveChan {
			//Problem: HAS TO GENERATE HUGE SLICE or run scan for each scan
			ch := sliceToChan(vCache)
			for urlI, depth := range recursive {
				wgT.Add(1)
				go requestWorker(ctx, logger, &wgT, urlI, ch, extensions, depth, responses, recursiveChan, tSizeP)
			}
		}
	}()
	go func() {
		wg.Wait()
		wgT.Wait()
		close(responses)
	}()
}

// To be edited
type Params struct {
	logger    *loggerCust
	w         io.Writer
	url       string
	keywords  chan string
	dict      string
	power     int
	responses chan Response
}

func bruteWebSite(url string, dict string, extensions []string, method string, power int, protocol string, verbose bool, w io.Writer) bool {
	power *= 10
	responses := make(chan Response, 5)
	keywords := make(chan string, 50)
	var logger loggerCust
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

	timer := time.Now()
	cw := CommonWriter{responses: responses, w: w, codes: []int{400, 500, 404}}

	logger.createLogger(url)
	url = addHTTPHTTPSProtocols(url, protocol)
	workerLauncher(ctx, cancel, &logger, w, url, keywords, dict, power, responses, sizeP, tSizeP, interrupt)

	go loader(500)

	checkColors(w)
	welcomeDataPrint(w, &logger, method, power, url, extensions)
	cw.writeWithColors(ctx, verbose)
	endDataPrint(w, size, tSize, time.Since(timer))

	logger.closeLogger()

	return true
}
