package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/idkwhyureadthis/practice/internal/models"
	"github.com/idkwhyureadthis/practice/internal/pkg/database"
)

var db = database.SetupDatabase()

func InitHandlers(srv chi.Router) {
	srv.Get("/parse", parseHandler)
	srv.Get("/get", getHandler)
}

func parseHandler(w http.ResponseWriter, r *http.Request) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var baseUrl = "https://api.hh.ru/vacancies"
	u, err := url.Parse(baseUrl)
	if err != nil {
		log.Println("failed to parse url", err)
		w.WriteHeader(500)
		w.Write([]byte{})
	}
	q := u.Query()
	if val := r.URL.Query().Get("salary"); val != "" {
		q.Add("salary", val)
		q.Add("currency", "RUR")
	}
	if val := r.URL.Query().Get("text"); val != "" {
		q.Add("text", val)
	}

	var firstPage models.PageData

	ur := baseUrl + "?" + q.Encode()
	resp, err := http.Get(ur)
	if err != nil {
		log.Println("failed to get:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println("failed to parse data from request", err)
		w.WriteHeader(500)
		w.Write([]byte{})
		return
	}

	err = json.Unmarshal(body, &firstPage)

	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte{})
		log.Println("failed marshaling JSON", err)
		return
	}

	rqPages := firstPage.Pages
	log.Println(rqPages)
	var pages models.PageData

	for i := 1; i <= rqPages; i++ {
		wg.Add(1)
		page := i
		query := url.Values{}
		for k, v := range q {
			query[k] = v
		}
		query.Set("page", fmt.Sprint(page))
		go func(query url.Values) {
			defer wg.Done()
			var jbs models.PageData
			u := baseUrl + "?" + query.Encode()
			resp, err := http.Get(u)
			if err != nil {
				log.Println("failed to get:", err)
				return
			}
			defer resp.Body.Close()
			data, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Println("failed to read data from response", err)
				return
			}
			err = json.Unmarshal(data, &jbs)
			if err != nil {
				log.Println("failed to scan in jbs", err)
				return
			}
			mu.Lock()
			pages.Jobs = append(pages.Jobs, jbs.Jobs...)
			pages.Pages = jbs.Pages
			mu.Unlock()
		}(query)
	}

	wg.Wait()

	body, err = json.Marshal(pages)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	db.SaveToDB(pages)
	w.WriteHeader(200)
	w.Write(body)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	city := r.URL.Query().Get("city")
	salary := r.URL.Query().Get("salary_from")
	experience := r.URL.Query().Get("experience")

	salaryInt, err := strconv.Atoi(salary)
	if err != nil {
		salaryInt = 0
	}

	experienceInt, err := strconv.Atoi(experience)
	if err != nil {
		experienceInt = 0
	}

	data := db.GetFromDB(name, city, salaryInt, experienceInt)
	body, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte{})
	}
	w.WriteHeader(200)
	w.Write(body)
}
