package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/barthr/newsapi"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Country struct {
	ID      int     `json:"id,omitempty"`
	Name    string  `json:"Name,omitempty"`
	Country string  `json:"country,omitempty"`
	Lat     float64 `json:"Lat,omitempty"`
	Long    float64 `json:"Long,omitempty"`
}

type News struct {
	ID          int    `json:"id,omitempty"`
	CountryID   int    `json:"countryId,omitempty"`
	Author      string `json:"author,omitempty"`
	Title       string `json:"Title,omitempty"`
	Description string `json:"description,omitempty"`
}

type countryInfo struct {
	ID          int     `json:"id,omitempty"`
	Name        string  `json:"Name,omitempty"`
	Lat         float64 `json:"Lat,omitempty"`
	Long        float64 `json:"Long,omitempty"`
	CountryID   int     `json:"countryId,omitempty"`
	Author      string  `json:"author,omitempty"`
	Title       string  `json:"Title,omitempty"`
	Description string  `json:"description,omitempty"`
}

type countryInformation struct {
	CountryInfor []countryInfo
}

var db *sql.DB

const (
	dbhost = "localhost"
	dbport = 5432
	dbuser = "postgres"
	dbpass = "pass"
	dbname = "insta"
)

func GetCountriesInfo(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	countryName := params["countryName"]

	// Fetch country information
	country := Country{}
	country = fetchCountry(country, countryName)
	if country == (Country{}) {
		fmt.Printf("no country found")
		return
	}

	// Fetch news if exists in db
	repos := countryInformation{}
	err := fetchNews(&repos, country)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	out, err := json.Marshal(repos)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Println(out)
	c := newsapi.NewClient("5f907a0ace284362aa4ea0b2e9dae1ff", newsapi.WithHTTPClient(http.DefaultClient))
	sources, _, err := c.GetSources(context.Background(), &newsapi.SourceParameters{
		Country: country.Country,
	})
	if err != nil {
		panic(err)
	}
	result := countryInformation{}
	for _, s := range sources.Sources {
		value := countryInfo{}
		value.Name = s.Name
		value.Title = s.Category
		value.Description = s.Description
		value.Lat = country.Lat
		value.Long = country.Long
		result.CountryInfor = append(result.CountryInfor, value)
	}
	fmt.Println(sources.Sources)
	json.NewEncoder(w).Encode(result)
}

func fetchCountry(country Country, countryName string) Country {
	query := fmt.Sprintf("SELECT id, name, country, lat, long FROM insta.country where name ='%s'", countryName)
	fmt.Println(query)
	rows, err := db.Query(query)
	if err != nil {
		return Country{}
	}
	defer rows.Close()
	for rows.Next() {
		value := Country{}
		err = rows.Scan(
			&value.ID,
			&value.Name,
			&value.Country,
			&value.Lat,
			&value.Long,
		)
		if err != nil {
			fmt.Println("error")
			return Country{}
		}
		fmt.Println(value)
		country.ID = value.ID
		country.Name = value.Name
		country.Lat = value.Lat
		country.Long = value.Long
		country.Country = value.Country
		return country
	}
	err = rows.Err()
	if err != nil {
		return Country{}
	}
	return Country{}
}

func fetchNews(info *countryInformation, country Country) error {
	query := fmt.Sprintf(`SELECT id, countryId, author, title, description FROM insta.news where countryId = %d`, country.ID)
	rows, err := db.Query(query)
	fmt.Println(query)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		value := countryInfo{}
		err = rows.Scan(
			&value.ID,
			&value.CountryID,
			&value.Title,
			&value.Description,
			&value.Author,
		)
		if err != nil {
			fmt.Println("error")
			return err
		}
		value.Lat = country.Lat
		value.Long = country.Long
		info.CountryInfor = append(info.CountryInfor, value)
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	return nil
}

func initDb() {
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		dbhost, dbport, dbuser, dbpass, dbname)

	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		fmt.Println("ghjkgjh")
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		fmt.Println("abcd")

		panic(err)
	}
}

func main() {
	router := mux.NewRouter()
	initDb()
	defer db.Close()
	router.HandleFunc("/countryInfo/{countryName}", GetCountriesInfo).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))
}
