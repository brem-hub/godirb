package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

//Structure for multiple extensions with -e flag
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

//Main struct for requests
type Response struct {
	url     string
	code    int
	size    int64
	keyword string
}

func (r *Response) Write() string {
	size := ByteCountIEC(r.size)
	str := fmt.Sprintf("%s  ::  %3d  ::  /%s/  ->  %s", size, r.code, r.keyword, r.url)
	return str
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

// Flag -clear: default clears log/ folder
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

func addHTTPHTTPSProtocols(url string, protocol string) string {
	switch {
	case protocol == "http":
		return "http://" + url
	case protocol == "https":
		return "https://" + url
	}
	return ""
}

func checkCodes(code int, codes []int) bool {
	for _, v := range codes {
		if v == code {
			return true
		}
	}
	return false
}

/*
Main func that does all the job
	url : 		url to attack
	dict: 		wordlist to use
	extensions: extensions to use [php, txt, etc], can be empty
	method: 	http method to use [get, post, head]
	goroutines: amount of goroutines to use, for simplicity * 10
	throttle:	throttiling - delay between requests
	depth:		depth of recursion [def: 1]
	protocol:	HTTP/HTTPS
	verbose:	print all data or only important
*/
func bruteWebSite(url string, dict string, extensions []string, method string, goroutines int, throttle int, depth int, protocol string, verbose bool, w io.Writer) bool {
	goroutines *= 10
	responses := make(chan Response, 5)
	keywords := make(chan string, 50)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	var (
		logger loggerCust
		size   int64
		tSize  int64
		tSizeP *int64
		sizeP  *int64
		om     outputManager
		rm     requestManager
		lm     loggerManager
	)

	sizeP = &size
	tSizeP = &tSize

	timer := time.Now()

	lm.New(url)
	lm.Start()

	logger = lm.GetLogger()

	url = addHTTPHTTPSProtocols(url, protocol)

	rm.New(w, &logger, url, method, keywords, dict, goroutines, depth, throttle, responses, extensions, sizeP, tSizeP)
	rm.Start(ctx, cancel)

	om.New(w, &logger, url, method, goroutines, responses, extensions, verbose, 500)
	om.Start(ctx, sizeP, tSizeP, timer)

	lm.Close()

	return true
}
