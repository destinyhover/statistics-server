package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Serving:", r.URL.Path, "from", r.Host)
	w.WriteHeader(http.StatusOK)
	body := "Thanks for visiting!\n"
	fmt.Fprintf(w, "%s", body)
}

func (a *App) deleteHandler(w http.ResponseWriter, r *http.Request) {
	paramStr := strings.Split(r.URL.Path, "/")
	fmt.Println("Path:", paramStr)
	if len(paramStr) < 3 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Not found:", r.URL.Path)
	}

	log.Println("Serving:", r.URL.Path, "from", r.Host)

	dataset := paramStr[2]
	a.mu.Lock()
	defer a.mu.Unlock()
	err := deleteEntry(dataset)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusNotFound)
		body := err.Error() + "\n"
		fmt.Fprintf(w, "%s", body)
		return
	}
	body := dataset + " deleted\n"
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", body)
}

func (a *App) listingHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Serving:", r.URL.Path, "from", r.Host)
	a.rwmu.RLock()
	body := list()
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", body)
	a.rwmu.RUnlock()
}

func (a *App) statusHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Serving:", r.URL.Path, "from", r.Host)
	a.rwmu.RLock()
	s := fmt.Sprintf("Total entries: %d\n", len(data))
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", s)
	a.rwmu.RUnlock()
}

func (a *App) insertHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Not enough arguments %s", parts)
		return
	}
	sarr := make([]float64, 0, len(parts))
	dataset := parts[3:]
	for _, v := range dataset {
		val, err := strconv.ParseFloat(v, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Cannot insert %f err:%s", val, err)
			return
		}
		sarr = append(sarr, val)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	entry := process(parts[2], sarr)
	err := insert(&entry)
	if err != nil {
		w.WriteHeader(http.StatusNotModified)
		fmt.Fprintf(w, "Failed in Inser() err:%s", err)
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Successfully added\n")
	}

	log.Println("Serving:", r.URL.Path, "from", r.Host)
}

func (a *App) searchHandler(w http.ResponseWriter, r *http.Request) {
	paramStr := strings.Split(r.URL.Path, "/")
	log.Println("Path:", paramStr)

	if len(paramStr) < 3 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Too short: %s\n", r.URL.Path)
	}

	dataSet := paramStr[2]
	a.mu.Lock()
	defer a.mu.Unlock()
	t := search(dataSet)
	if t == nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Couldn't ve found: %s\n", dataSet)
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s %d %f %f\n", t.Name, t.Len, t.Mean, t.StdDev)
	}
	log.Println("Serving:", r.URL.Path, "from", r.Host)

}
