package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
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

type CotacaoResponse struct {
	Cotacao string `json:"cotacao"`
}

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
	var cotacao = CotacaoResponse{}
	cotacao.Cotacao = awesomeApiUSDBRL.Usdbrl.Bid
	currentDollarExchangeRate, err := strconv.ParseFloat(awesomeApiUSDBRL.Usdbrl.Bid, 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = save(db, currentDollarExchangeRate)
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
	json.NewEncoder(w).Encode(cotacao)
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
	const createTable string = `CREATE TABLE IF NOT EXISTS dollar_exchange_rates (
	id INTEGER NOT NULL PRIMARY KEY,
	date DATETIME NOT NULL,
	dollar_exchange_rate REAL
	);`
	db.Exec(createTable)
	return db, nil
}

func save(db *sql.DB, currentDollarExchangeRate float64) error {
	const insertCommand = `INSERT INTO dollar_exchange_rates(date, dollar_exchange_rate) 
	VALUES(datetime('now', 'localtime'), ?)`
	stmt, err := db.Prepare(insertCommand)
	if err != nil {
		return err
	}
	defer stmt.Close()
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()
	_, err = stmt.ExecContext(ctx, currentDollarExchangeRate)
	if err != nil {
		return err
	}
	return nil
}
