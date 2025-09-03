package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"slices"
	"sync"
	"time"
)

type App struct {
	rwmu sync.RWMutex
}

func getenv(k, def string) string {
	if v := os.Getenv((k)); v != "" {
		return v
	}
	return def
}

var JSONFILE = getenv("DATA_PATH", "./data/data.json")
var PORT = getenv("PORT", ":1234")

type PhoneBook []Entry

var data = PhoneBook{}
var index map[string]int

type Entry struct {
	Name    string
	Len     int
	Minimum float64
	Maximum float64
	Mean    float64
	StdDev  float64
}

func process(file string, values []float64) Entry {
	currentEntry := Entry{}
	currentEntry.Name = file

	currentEntry.Len = len(values)
	currentEntry.Minimum = slices.Min(values)
	currentEntry.Maximum = slices.Max(values)

	meanValue, standardDeviation := stdDev(values)
	currentEntry.Mean = meanValue
	currentEntry.StdDev = standardDeviation
	return currentEntry
}

func stdDev(x []float64) (float64, float64) {
	sum := float64(0)
	for _, val := range x {
		sum += val
	}
	meanValue := sum / float64(len(x))

	var squared float64
	for i := 0; i < len(x); i++ {
		squared = squared + math.Pow((x[i]-meanValue), 2)
	}

	standardDeviation := math.Sqrt(squared / float64(len(x)))
	return meanValue, standardDeviation
}

// ============= json funcs
func DeSerialize(slice interface{}, r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(slice)
}

func Serialize(slice interface{}, w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(slice)
}

func saveJSONFile(filepath string) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	err = Serialize(&data, f)
	if err != nil {
		return err
	}
	return nil
}

func readJSONFile(filepath string) error {
	_, err := os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			_, err = os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE, 0644)
			return err
		}
		return err
	}

	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	err = DeSerialize(&data, f)
	if err != nil {
		return err
	}
	return nil
}

func createIndex() {
	index = make(map[string]int)
	for i, v := range data {
		key := v.Name
		index[key] = i
	}
}

func insert(pS *Entry) error {
	_, ok := index[pS.Name]
	if ok {
		return fmt.Errorf("Already exists %s", pS.Name)
	}

	data = append(data, *pS)
	createIndex()
	err := saveJSONFile(JSONFILE)
	if err != nil {
		return err
	}
	return nil
}

func deleteEntry(key string) error {
	i, ok := index[key]
	if !ok {
		return fmt.Errorf("Does not exist: %s", key)
	}
	data = append(data[:i], data[i+1:]...)
	delete(index, key)

	err := saveJSONFile(JSONFILE)
	if err != nil {
		return err
	}
	return nil
}

func search(key string) *Entry {
	i, ok := index[key]
	if !ok {
		return nil
	}
	return &data[i]
}

func list() string {
	var all string
	for _, v := range data {
		all = all + fmt.Sprintf("%s\t%d\t%f\t%f\n", v.Name, v.Len, v.Mean, v.StdDev)
	}
	return all
}

func main() {
	err := readJSONFile(JSONFILE)
	if err != nil && err != io.EOF {
		fmt.Println("Error:", err)
		return
	}

	createIndex()
	mux := http.NewServeMux()
	s := &http.Server{
		Addr:         PORT,
		Handler:      mux,
		IdleTimeout:  10 * time.Second,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	}

	a := &App{}
	mux.Handle("/", http.HandlerFunc(defaultHandler))
	mux.Handle("/list", http.HandlerFunc(a.listingHandler))
	mux.Handle("/search", http.HandlerFunc(a.searchHandler))
	mux.Handle("/search/", http.HandlerFunc(a.searchHandler))
	mux.Handle("/status", http.HandlerFunc(a.statusHandler))
	mux.Handle("/delete/", http.HandlerFunc(a.deleteHandler))
	mux.HandleFunc("/insert", a.insertHandler)
	mux.HandleFunc("/insert/", a.insertHandler)
	fmt.Println("Ready to serve at", PORT)
	err = s.ListenAndServe()
	if err != nil {
		fmt.Println(err)
		return
	}
}
