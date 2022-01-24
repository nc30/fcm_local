package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	flag.StringVar(&addr, "addr", ":3000", "listening addr")
	flag.Parse()
}

var addr = ""
var buffer [][]byte

type response struct {
	Count int      `json:"count"`
	Datas []string `json:"datas"`
}

type fcmPayload struct {
	To               string `json:"to"`
	Priority         string `json:"priority"`
	ContentAvailable bool   `json:"content_available"`
	TimeToLive       int    `json:"time_to_live"`
	Data             struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	Notification struct {
		Badge int    `json:"badge"`
		Title string `json:"title"`
		Body  string `json:"body"`
	}
}

type fcmResponse struct {
	MulticastId  int64       `json:"multicast_id"`
	Success      int         `json:"success"`
	Failure      int         `json:"failure"`
	CanonicalIds int         `json:"canonical_ids"`
	Results      []fcmResult `json:"results"`
}
type fcmResult struct {
	MessageId string `json:"message_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

func main() {
	buffer = make([][]byte, 0)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		resp := &response{
			Count: 0,
			Datas: make([]string, 0),
		}

		for _, b := range buffer {
			resp.Count++
			resp.Datas = append(resp.Datas, string(b))
		}

		re, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Add("Content-Type", "application/json;charset=utf-8")
			w.Write([]byte(err.Error()))
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json;charset=utf-8")
		w.Write(re)
	})

	r.Get("/reset", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json;charset=utf-8")
		w.Write([]byte("{}"))

		buffer = make([][]byte, 0)
	})

	r.Post("/fcm/send", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		payload := new(fcmPayload)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Add("Content-Type", "application/json;charset=utf-8")
			w.Write([]byte(err.Error()))
		}

		log.Println(string(body))

		json.Unmarshal(body, payload)
		buffer = append(buffer, body)

		var result *fcmResponse
		switch {
		case len(payload.To) > 3 && payload.To[:4] == "fail":
			result = &fcmResponse{
				MulticastId:  rand.Int63n(0xffffffff),
				Success:      0,
				Failure:      1,
				CanonicalIds: 0,
				Results: []fcmResult{
					fcmResult{
						Error: "NotRegistered",
					},
				},
			}
		default:
			result = &fcmResponse{
				MulticastId:  rand.Int63n(0xffffffff),
				Success:      1,
				Failure:      0,
				CanonicalIds: 0,
				Results: []fcmResult{
					fcmResult{
						MessageId: strconv.FormatInt(rand.Int63n(0xffffffff), 10),
					},
				},
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json;charset=utf-8")

		z, _ := json.Marshal(result)
		w.Write(z)
	})

	fmt.Fprintf(os.Stdout, "listening %s \r\n", addr)
	err := http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatal(err)
	}
}
