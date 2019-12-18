package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime/debug"
	"strconv"

	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

func main() {
	err := http.ListenAndServe(":3000", recoverMw(WebHandler()))
	if err != nil {
		return
	}
}

//WebHandler handles web request
func WebHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/", sourceCodeHandler)
	mux.HandleFunc("/panic/", panicDemo)
	mux.HandleFunc("/panic-after/", panicAfterDemo)
	mux.HandleFunc("/", hello)
	return mux
}

//sourceCodeHandler handles request by retriving form value for path variable and display in chrome
func sourceCodeHandler(w http.ResponseWriter, r *http.Request) {
	path := r.FormValue("path")
	lineStr := r.FormValue("line")
	line, err := strconv.Atoi(lineStr)
	file, err := os.Open(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b := bytes.NewBuffer(nil)
	io.Copy(b, file)
	var lines [][2]int
	if line > 0 {
		lines = append(lines, [2]int{line, line})
	}
	lexer := lexers.Get("go")
	iterator, err := lexer.Tokenise(nil, b.String())
	style := styles.Get("github")
	formatter := html.New(html.TabWidth(2), html.WithLineNumbers(), html.LineNumbersInTable(), html.HighlightLines(lines))
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<style>pre { font-size: 1.2em; }</style>")
	formatter.Format(w, style, iterator)
}

//recoverMw used as recovery middleware to handle panic occurs
func recoverMw(app http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println(err)
				stack := debug.Stack()
				log.Println(string(stack))
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "<h1> Panic: %v</h1><pre>%s</pre>", err, makeLinks(string(stack)))
			}
		}()

		app.ServeHTTP(w, r)
	}
}

//panicDemo calls panic function only
func panicDemo(w http.ResponseWriter, r *http.Request) {
	funcThatPanics()
}

//panicAfterDemo prints Hello string with panic msg
func panicAfterDemo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Hello!</h1>")
	funcThatPanics()
}

//funcThatPanics is responsible to generate panic
func funcThatPanics() {
	panic("Oh no!")
}

//hello prints default web page with hello string only
func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<h1>Hello!</h1>")
}

//makeLinks  takes the string and form the links at the line numbers
func makeLinks(stack string) string {
	re := regexp.MustCompile("\t.*:[0-9]*")
	lines := re.FindAllString(stack, -1)

	re = regexp.MustCompile(":")
	for _, line := range lines {
		splits := re.Split(line, -1)
		link := "<a href='/debug?path=" + splits[0] + "&line=" + splits[1] + "'>" + line + "</a>"
		reg := regexp.MustCompile(line)
		stack = reg.ReplaceAllString(stack, link)
	}
	return stack
}
