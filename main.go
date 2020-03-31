package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

type alarmData struct {
	PhoneNumber   string
	Time          time.Time
	NumberOfCalls int
}

type indexData struct {
	ServerTime string
	Message    string
	Alarms     []alarmData
}

var alarms []alarmData
var tpl = template.Must(template.ParseFiles("index.html"))

func (a *alarmData) exec() {
	go func() {
		sendCallUntil(a.PhoneNumber, a.NumberOfCalls, a.Time)
		for i, v := range alarms {
			if v == *a {
				alarms = append(alarms[0:i], alarms[i+1:]...)
			}
		}
	}()
}

func (a *alarmData) exceededLimit() bool {
	var count int
	for _, v := range alarms {
		if v.PhoneNumber == a.PhoneNumber {
			count++
		}
	}
	return count >= 3
}

func setHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	} else if r.Method != "POST" || r.URL.Path != "/set" {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	alarm := alarmData{}
	matched, err := regexp.Match(`^\+\d{5,15}$`, []byte(r.FormValue("phone_number")))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !matched {
		http.Redirect(w, r, "/?msg=Invalid phone number.", http.StatusTemporaryRedirect)
		return
	}
	alarm.PhoneNumber = r.FormValue("phone_number")
	now := time.Now()
	alarmTime, err := time.ParseInLocation("15:04", r.FormValue("time"),
		now.Location())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	alarmTime = alarmTime.AddDate(now.Year(), int(now.Month())-1, now.Day()-1)
	if alarmTime.Sub(now) <= 0 {
		alarmTime = alarmTime.AddDate(0, 0, 1)
	}
	alarm.Time = alarmTime
	numberOfCalls, err := strconv.Atoi(r.FormValue("noc"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if numberOfCalls <= 0 {
		http.Redirect(w, r, "/?msg=Invalid number of calls.", http.StatusTemporaryRedirect)
		return
	}
	alarm.NumberOfCalls = numberOfCalls
	if alarm.exceededLimit() {
		http.Redirect(w, r, "/?msg=Limit exceeded.", http.StatusTemporaryRedirect)
		return
	}
	alarms = append(alarms, alarm)
	alarm.exec()
	log.Printf("added job %v * %v @ %v", alarm.NumberOfCalls, alarm.PhoneNumber, alarm.Time)
	http.Redirect(w, r, "/?msg=Alarm set successfully.", http.StatusTemporaryRedirect)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 page not found.", http.StatusNotFound)
		return
	} else if err := r.ParseForm(); err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	tpl.Execute(w, indexData{
		ServerTime: time.Now().String(),
		Message:    r.FormValue("msg"),
		Alarms:     alarms,
	})
}

func main() {
	fmt.Println("bangunin \u2014 phone call-based alarm")
	port := flag.Int("port", 7167, "port of web server")
	flag.Parse()

	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	mux.HandleFunc("/set", setHandler)
	mux.HandleFunc("/", indexHandler)

	log.Printf("serving on :%v", *port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), mux))
}
