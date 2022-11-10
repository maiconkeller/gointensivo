package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/maiconkeller/gointensivo/internal/order/infra/database"
	"github.com/maiconkeller/gointensivo/internal/order/usecase"
	"github.com/maiconkeller/gointensivo/pkg/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"

	//sqlite3
	_ "github.com/mattn/go-sqlite3"
)

//Thread 1
func main() {
	println("consumer initiated...")

	db, err := sql.Open("sqlite3", "./orders.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	repository := database.NewOrderRepository(db)
	uc := usecase.NewCalculateFinalPriceUseCase(repository)

	ch, err := rabbitmq.OpenChannel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	out := make(chan amqp.Delivery) //channel

	go rabbitmq.Consume(ch, out) //Thread 2

	qtdWorkers := 150
	for i := 0; i < qtdWorkers; i++ {
		go worker(out, uc, i)
	}

	// criando um server http para retornar o total de msgs lidas
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tUC := usecase.NewGetTotalUseCase(repository)
		total, err := tUC.Execute()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		json.NewEncoder(w).Encode(total)
	})

	http.ListenAndServe(":8080", nil)

}

func worker(deliveryMessage <-chan amqp.Delivery, uc *usecase.CalculateFinalPriceUseCase, workerID int) {
	for msg := range deliveryMessage {
		var inputDTO usecase.OrderInputDTO

		err := json.Unmarshal(msg.Body, &inputDTO)
		if err != nil {
			panic(err)
		}
		outputDTO, err := uc.Execute(inputDTO)
		if err != nil {
			panic(err)
		}
		msg.Ack(false)
		fmt.Printf("worker %d has processed order %s\n", workerID, outputDTO.ID)
		time.Sleep(1 * time.Second)
	}
}
