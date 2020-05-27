package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
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

type outputManager struct {
	cw         CommonWriter
	speed      time.Duration
	logger     *loggerCust
	method     string
	gorutines  int
	target     string
	verbose    bool
	extensions []string
}

func (om *outputManager) New(w io.Writer, logger *loggerCust, target string, method string, gorutines int, responses chan Response, extensions []string, verbose bool, speed time.Duration) {
	om.cw = CommonWriter{responses: responses, w: w, codes: []int{400, 500, 404}}
	om.speed = speed
	om.logger = logger
	om.method = method
	om.gorutines = gorutines
	om.target = target
	om.extensions = extensions
	om.verbose = verbose
}

func (om *outputManager) Start(ctx context.Context, size *int64, tSize *int64, timer time.Time) {
	go om.loader()
	om.checkColors()
	om.welcomeDataPrint()
	om.cw.writeWithColors(ctx, om.verbose)
	om.endDataPrint(*size, *tSize, time.Since(timer))
}

/*
Create loader animation.
*/
func (om *outputManager) loader() {
	time.Sleep(150 * time.Millisecond)
	for {
		fmt.Printf("\r\\")
		time.Sleep(om.speed * time.Millisecond)
		fmt.Printf("\r/")
		time.Sleep(om.speed * time.Millisecond)
	}
}

/*
If writer is file - turn Colors off
*/
func (om *outputManager) checkColors() {
	if om.cw.w != os.Stdout {
		Cyan.DisableColor()
		Blue.DisableColor()
		Red.DisableColor()
		Yellow.DisableColor()
		Green.DisableColor()
		GreenBg.DisableColor()
	}
}

func (om *outputManager) welcomeDataPrint() {
	if len(om.extensions) == 0 {
		om.extensions = append(om.extensions, "none")
	}

	Blue.Fprintln(om.cw.w, "_________     _____________       ______\n__  ____/________  __ \\__(_)_________  /_\n_  / __ _  __ \\_  / / /_  /__  ___/_  __ \\\n/ /_/ / / /_/ /  /_/ /_  / _  /   _  /_/ /\n\\____/  \\____//_____/ /_/  /_/    /_.___/")
	fmt.Fprintln(om.cw.w)
	fmt.Fprintf(om.cw.w, "%s %s %s %s %s %s %s %s\n\n", Blue.Sprint("HTTP Method:"), Green.Sprint(om.method), Yellow.Sprint("|"),
		Blue.Sprint("Gorutines:"), Green.Sprint(om.gorutines), Yellow.Sprint("|"),
		Blue.Sprint("Extensions:"), Green.Sprint(strings.Join(om.extensions, " ")))
	fmt.Fprintf(om.cw.w, "%s %s\n\n", Blue.Sprint("Error log:"), Green.Sprint(om.logger.file.Name()))
	fmt.Fprintf(om.cw.w, "%s %s\n\n", Blue.Sprint("Target:"), Green.Sprint(om.target))
	Blue.Fprintln(om.cw.w, ":::Starting:::")
	fmt.Fprintln(om.cw.w, "+---------------+")
}

func (om *outputManager) endDataPrint(wordsize int64, donesize int64, elapsedTime time.Duration) {
	fmt.Fprintln(om.cw.w, "\r+---------------+")
	Blue.Fprintln(om.cw.w, ":::Completed:::")
	fmt.Fprintln(om.cw.w)
	fmt.Fprintf(om.cw.w, "%s/%s codes returned\n", Green.Sprint(donesize), Blue.Sprint(wordsize))
	fmt.Fprintf(om.cw.w, "Elapsed time: %s\n", Green.Sprint(elapsedTime))
}

type CommonWriter struct {
	responses chan Response
	codes     []int
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
				fmt.Fprintln(r.w)
			} else {
				if !checkCodes(response.code, r.codes) {
					color.Fprintf(r.w, "\r%s", response.Write())
					fmt.Fprintln(r.w)
					size += tmp
				}
			}
		}
	}
	return size, nil
}
