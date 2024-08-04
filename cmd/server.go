package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/andretefras/fullcycle-challenge-1/internal/exchangerate"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"time"
)

func getDollarRate(ctx context.Context) (*exchangerate.ExchangeRate, error) {
	log.Println("Criando requisição HTTP para a API de cotação.")
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		log.Printf("Erro ao criar a requisição: %v", err)
		return nil, err
	}

	client := &http.Client{}
	log.Println("Enviando requisição para a API de cotação.")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Erro ao obter resposta da API: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]map[string]string
	log.Println("Decodificando resposta JSON da API.")
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Erro ao decodificar JSON: %v", err)
		return nil, err
	}

	return &exchangerate.ExchangeRate{Bid: result["USDBRL"]["bid"]}, nil
}

func exchangeRateHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Recebido pedido de cotação via HTTP.")
		ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
		defer cancel()

		log.Println("Consultando cotação atual do dólar.")
		rate, err := getDollarRate(ctx)
		if err != nil {
			log.Printf("Erro ao buscar a cotação do dólar: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		ctxDB, cancelDB := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancelDB()

		log.Println("Persistindo cotação no banco de dados.")
		_, err = db.ExecContext(ctxDB, "INSERT INTO rates (bid) VALUES (?)", rate.Bid)
		if err != nil {
			log.Printf("Erro ao salvar no banco de dados: %v", err)
			http.Error(w, "Database timeout", http.StatusInternalServerError)
			return
		}

		log.Println("Cotação persistida e sendo enviada ao cliente.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rate)
	}
}

func initializeDB(db *sql.DB) {
	log.Println("Inicializando banco de dados e tabela de cotações.")
	createTableSQL := `CREATE TABLE IF NOT EXISTS rates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		bid TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(createTableSQL); err != nil {
		log.Fatalf("Falha ao criar a tabela: %v", err)
	}
	log.Println("Tabela de cotações pronta para uso.")
}

func main() {
	log.Println("Abrindo conexão com banco de dados SQLite.")
	db, err := sql.Open("sqlite3", "./rates.db")
	if err != nil {
		log.Fatalf("Erro ao abrir banco de dados: %v", err)
	}
	defer db.Close()

	initializeDB(db)

	log.Println("Configurando rota HTTP /cotacao e iniciando servidor.")
	http.HandleFunc("/cotacao", exchangeRateHandler(db))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
