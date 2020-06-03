package main

import (
	"flag"
	"fmt"
	"os"
)

/*TODO:
create POST
*/
var (
	url        = flag.String("u", "", "specify url to run. Usage: -u <url>")
	customDict = flag.String("cd", "", "use custom dictionary. Usage: -cd <path>")
	verbose    = flag.Bool("v", false, "show more information(each request)")
	method     = flag.String("m", "get", "specify method to use [get, post, head]")
	file       = flag.String("f", "", "specify file to output into. Usage: -f <path>")
	power      = flag.Int("go", 1, "amount of goroutines. Usage -p [1...5]")
	protocol   = flag.String("p", "https", "specify protocol. Usage: -protocol <http/https>")
	clear      = flag.Bool("clear", false, "use this flag to clear log/ folder.")
	recursive  = flag.Int("d", 1, "specify depth of recursion. 1 equals <url>/root/")
	throttle   = flag.Int("t", 0, "specify delay between requests in sec.")
	extensions StringSlice
)

func main() {
	dict := "data/dicc.txt"
	flag.Var(&extensions, "e", "extensions to pass. Usage: -e=php,txt,rcc")
	flag.Parse()
	if *power > 5 {
		fmt.Printf("Do you really want to use %s goroutines?\n", Red.Sprint(*power*10))
		os.Exit(1)
	}
	if *clear {
		err := clearDir("log/")
		if err != nil {
			Red.Println(err)
		}
		Green.Println("Log folder has been cleared")
		os.Exit(1)
	}
	if *method != "post" && *method != "get" && *method != "head" {
		fmt.Printf("%s is not a valid http method. Use [get, post, head]\n", Red.Sprint(*method))
		os.Exit(1)
	}
	if *url == "" {
		fmt.Println("Specify url to run, usage: -u <url>")
		os.Exit(1)
	}
	if *protocol != "http" && *protocol != "https" {
		fmt.Println("Protocol should be HTTP or HTTPS")
		os.Exit(1)
	}
	if *customDict != "" {
		dict = *customDict
	}
	if *file == "" {
		bruteWebSite(*url, dict, extensions, *method, *power, *throttle, *recursive, *protocol, *verbose, os.Stdout)
	} else {
		file, err := os.Create(*file)
		if err != nil {
			Red.Println(err)
			os.Exit(1)
		}
		bruteWebSite(*url, dict, extensions, *method, *power, *throttle, *recursive, *protocol, *verbose, file)

	}
}
