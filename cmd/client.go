package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/andretefras/fullcycle-challenge-1/internal/exchangerate"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	log.Println("Iniciando o cliente...")
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	log.Println("Criando a requisição HTTP...")
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatalf("Erro ao criar a requisição: %v", err)
	}

	log.Println("Enviando a requisição HTTP...")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Erro ao fazer a requisição: %v", err)
	}
	defer resp.Body.Close()

	log.Println("Lendo o corpo da resposta...")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Erro ao ler o corpo da resposta: %v", err)
	}

	var cotacao exchangerate.ExchangeRate
	log.Println("Fazendo o parsing do JSON...")
	err = json.Unmarshal(body, &cotacao)
	if err != nil {
		log.Fatalf("Erro ao fazer o parsing do JSON: %v", err)
	}

	log.Println("Criando o arquivo...")
	file, err := os.Create("cotacao.txt")
	if err != nil {
		log.Fatalf("Erro ao criar o arquivo: %v", err)
	}
	defer file.Close()

	log.Printf("Escrevendo no arquivo: Dólar: %s", cotacao.Bid)
	if _, err = file.WriteString(fmt.Sprintf("Dólar: %s", cotacao.Bid)); err != nil {
		log.Fatalf("Erro ao escrever no arquivo: %v", err)
	}
}
