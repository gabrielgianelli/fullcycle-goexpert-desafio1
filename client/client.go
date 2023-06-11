package main

import (
	"context"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type Cotacao struct {
	Dolar struct {
		Valor string `json:"bid"`
	} `json:"USDBRL"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	var cotacao Cotacao
	err = json.Unmarshal(body, &cotacao)
	if err != nil {
		panic(err)
	}
	file, err := os.Create("cotacao.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	cotacaoTemplate := template.Must(template.New("CotacaoTemplate").Parse("Dólar: {{.Valor}}"))
	err = cotacaoTemplate.Execute(file, cotacao.Dolar)
	if err != nil {
		panic(err)
	}
}
