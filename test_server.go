package main

import (
	"fmt" // пакет для форматированного ввода вывода
	// пакет для логирования
	"net/http" // пакет для поддержки HTTP протокола
	// пакет для работы с  UTF-8 строками
)

func Home(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, "Hello GoTest!") }
func Kek(w http.ResponseWriter, r *http.Request)  { fmt.Fprintf(w, "Kek") }
func Lol(w http.ResponseWriter, r *http.Request)  { fmt.Fprintf(w, "Lol") }
func C200(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }
func C302(w http.ResponseWriter, r *http.Request) { w.WriteHeader(302) }
func C301(w http.ResponseWriter, r *http.Request) { w.WriteHeader(301) }
func C307(w http.ResponseWriter, r *http.Request) { w.WriteHeader(307) }
func C308(w http.ResponseWriter, r *http.Request) { w.WriteHeader(308) }

func main() {
	http.HandleFunc("/", C200)
	http.HandleFunc("/favicon.ico", http.NotFound)
	http.HandleFunc("/test301", C301)
	http.HandleFunc("/test302", C302)

	http.HandleFunc("/kek", C200)
	http.HandleFunc("/lol", C200)
	http.HandleFunc("/user", C200)
	http.HandleFunc("/login", C200)
	http.HandleFunc("/feed", C200)

	http.HandleFunc("/data", C307)
	http.HandleFunc("/admin", C307)
	http.HandleFunc("/test307", C307)
	http.HandleFunc("/test308", C308)
	http.HandleFunc("/app", C308)
	http.HandleFunc("/calculator", C308)
	http.HandleFunc("/checkext", C307)
	http.HandleFunc("/checkext.php", C307)
	http.HandleFunc("/checkext.txt", C307)

	http.HandleFunc("/search", http.NotFound)
	http.HandleFunc("/kok", http.NotFound)
	http.HandleFunc("/check", http.NotFound)
	http.HandleFunc("/cheburek", http.NotFound)
	http.HandleFunc("/dod", http.NotFound)

	http.ListenAndServe(":9000", nil)
}
