package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type AwesomeApiUSDBRL struct {
	Usdbrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

// type exchangeRates struct {
// 	ID         int `gorm:"primarykey"`
// 	Code       string
// 	Codein     string
// 	Name       string
// 	High       float64
// 	Low        float64
// 	VarBid     float64
// 	PctChange  float64
// 	Bid        float64
// 	Ask        float64
// 	Timestamp  int64
// 	CreateDate time.Time
// 	gorm.Model
// }

// type CotacaoResponse struct {
// 	Cotacao string `json:"cotacao"`
// }

func main() {
	http.HandleFunc("/cotacao", cotacaoHandler)
	http.ListenAndServe(":8080", nil)
}

func cotacaoHandler(w http.ResponseWriter, r *http.Request) {
	awesomeApiUSDBRL, err := dollarExchangeRate()
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			w.WriteHeader(http.StatusGatewayTimeout)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	db, err := newDatabaseConnection()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()
	err = save(db, awesomeApiUSDBRL)
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			w.WriteHeader(http.StatusGatewayTimeout)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		println(err.Error())
		return
	}
	w.Header().Set("Content-Type", "Application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(awesomeApiUSDBRL)
}

func dollarExchangeRate() (*AwesomeApiUSDBRL, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET",
		"https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var awesomeApiUSDBRL AwesomeApiUSDBRL
	err = json.Unmarshal(body, &awesomeApiUSDBRL)
	if err != nil {
		return nil, err
	}
	return &awesomeApiUSDBRL, nil
}

func newDatabaseConnection() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "goexpert_desafio1.sqlite")
	if err != nil {
		return nil, err
	}
	const createTable string = `CREATE TABLE IF NOT EXISTS exchange_rates (
	id INTEGER NOT NULL PRIMARY KEY,
	code TEXT,
    codein TEXT,
    name TEXT,
    high REAL,
    low REAL,
    varBid REAL,
    pctChange REAL,
    bid REAL,
    ask REAL,
    timestamp TEXT,
    create_date TEXT
	);`
	db.Exec(createTable)
	return db, nil
}

func save(db *sql.DB, api *AwesomeApiUSDBRL) error {
	const insertCommand = `INSERT INTO exchange_rates(
		code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, create_date) 
	VALUES(
		?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	stmt, err := db.Prepare(insertCommand)
	if err != nil {
		return err
	}
	defer stmt.Close()
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()
	_, err = stmt.ExecContext(ctx, api.Usdbrl.Code, api.Usdbrl.Codein, api.Usdbrl.Name,
		api.Usdbrl.High, api.Usdbrl.Low, api.Usdbrl.VarBid, api.Usdbrl.PctChange, api.Usdbrl.Bid,
		api.Usdbrl.Ask, api.Usdbrl.Timestamp, api.Usdbrl.CreateDate)
	if err != nil {
		return err
	}
	return nil
}
