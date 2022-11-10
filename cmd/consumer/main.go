package main

import (
	"database/sql"
	"encoding/json"
	"fmt"

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
	go rabbitmq.Consume(ch, out)    //Thread 2

	for msg := range out {
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
		fmt.Println(outputDTO) //Thread 1
	}

}
