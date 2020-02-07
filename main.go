package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/shyam81992/Weather-Monster-job/models"

	"github.com/shyam81992/Weather-Monster-job/config"
	"github.com/shyam81992/Weather-Monster-job/db"
	"github.com/shyam81992/Weather-Monster-job/helper"
	"github.com/streadway/amqp"
)

func initialize() {

	forever := make(chan bool)

	c := make(chan *amqp.Error)
	go func() {
		err := <-c
		log.Println("reconnect: " + err.Error())
		time.Sleep(time.Duration(60) * time.Second)
		initialize()
		forever <- true
		
	}()

	conn, err := amqp.Dial(config.RabbitConfig["uri"])
	if err != nil {
		fmt.Println(err, "Failed to connect to RabbitMQ")
		return;
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		fmt.Println(err, "Failed to register a consumer")
		return;
	}
	defer ch.Close()

	conn.NotifyClose(c)
	q, err := ch.QueueDeclare(
		config.RabbitConfig["queuename"], // name
		true,                             // durable
		false,                            // delete when unused
		false,                            // exclusive
		false,                            // no-wait
		nil,                              // arguments
	)
	if err != nil {
		fmt.Println(err, "Failed to register a consumer")
		return;
	}
	QueueListen(ch, q.Name)
	<-forever
}

func QueueListen(ch *amqp.Channel, name string) {
	msgs, err := ch.Consume(
		name,  // queue
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		fmt.Println(err, "Failed to register a consumer")
		// close the channel and handle the channel close event
		return;
	}

	go func() {
		for d := range msgs {
			fmt.Println("Received a message:", string(d.Body))
			processmsg(d.Body)
			log.Printf("Done")
			d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

}

func processmsg(msg []byte) {

	var temp models.Temperature

	if err := json.Unmarshal(msg, &temp); err != nil {
		fmt.Println(err, "Invalid msg", string(msg))
	}

	var webhooks []string

	continueloop := true
	size := 0
	var lastprocessedid int64
	for continueloop {
		var id int64
		ctx, _ := context.WithTimeout(context.Background(), 1*time.Minute)
		rows, err := db.Db.QueryContext(ctx, `select id,callback_url from webhook where city_id=$1
		 and id > $2 limit 30000`, temp.CityID, lastprocessedid)
		if err != nil {
			// handle this error better than this
			panic(err)
		}
		defer rows.Close()
		for rows.Next() {

			var webhook string
			err = rows.Scan(&id, &webhook)
			if err != nil {
				// handle this error
				panic(err)
			}
			webhooks = append(webhooks, webhook)

		}
		// get any error encountered during iteration
		err = rows.Err()
		if err != nil {
			panic(err)
		}

		if size == len(webhooks) {
			continueloop = false
		} else {
			size = len(webhooks)
			lastprocessedid = id
		}

	}

	for _, wh := range webhooks {
		err := helper.PostDataToWM(wh, msg)
		if err != nil {
			fmt.Println(err, "PostDataToWM")
			// send error notification or mail or implement dlx retry or retrylogic at code level
		}
	}

}

func main() {
	config.LoadConfig()
	db.InitDb()

	initialize()

}
