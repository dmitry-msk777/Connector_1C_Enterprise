package connector

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/streadway/amqp"

	"encoding/json"

	rootsctuct "github.com/dmitry-msk777/Connector_1C_Enterprise/rootdescription"
)

var ConnectorV Connector

type Connector struct {
	DataBaseType     string
	RabbitMQ_channel *amqp.Channel
	Global_settings  rootsctuct.Global_settings
	LoggerCRM        rootsctuct.LoggerCRM
}

func (Connector *Connector) SetSettings(Global_settings rootsctuct.Global_settings) error {

	Connector.DataBaseType = Global_settings.DataBaseType
	if Global_settings.DataBaseType == "" {
		Connector.DataBaseType = "DemoRegime"
		Global_settings.DataBaseType = "DemoRegime"
	}

	Connector.Global_settings = Global_settings

	err := Connector.InitDataBase()
	if err != nil {
		return err
	}

	return nil

}

func (Connector *Connector) InitDataBase() error {

	if Connector.Global_settings.UseRabbitMQ {
		Connector.InitRabbitMQ()
		//go RabbitMQ_Consumer()
	}

	return nil
}

func (Connector *Connector) ConsumeFromQueue() (map[string]rootsctuct.Customer_struct, error) {

	if Connector.RabbitMQ_channel == nil {
		err := errors.New("Connection to RabbitMQ not established")
		return nil, err
	}

	var customer_map_json = make(map[string]rootsctuct.Customer_struct)

	q, err := Connector.RabbitMQ_channel.QueueDeclare(
		"Customer___add_change", // name
		false,                   // durable
		false,                   // delete when unused
		false,                   // exclusive
		false,                   // no-wait
		nil,                     // arguments
	)

	if err != nil {
		fmt.Println("Failed to declare a queue: ", err)
		return nil, err
	}

	msgs, err := Connector.RabbitMQ_channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if err != nil {
		fmt.Println("Failed to register a consumer: ", err)
		return nil, err
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// 1. Для промышленного использования рекомендую запустить горутину, где
	// получать данные из канала методом range передавая эти данные в
	// опубликованный в 1С http сервис.
	// 2. В процедуру добавлено замедление т.к на тестовом RabbitMQ наблюдались перебои
	// с получением сооббщений, сейчас все ок
	// 3. Так же я вынужден разрывать соединение, чтобы при завершении на строне RabbitMQ
	// не отставалось потребителя, возможно это можно сделать какой-то функцией, убить потребитель
	// после завершения, а не убивать все соединение.

	go func() {
		for {
			time.Sleep(1000 * time.Millisecond)
			select {
			case msg := <-msgs:
				fmt.Println("msg = ", msg)
				Customer_struct := rootsctuct.Customer_struct{}

				err = json.Unmarshal(msg.Body, &Customer_struct)
				if err != nil {
					Connector.LoggerCRM.ErrorLogger.Println(err.Error())
				}

				customer_map_json[Customer_struct.Customer_id] = Customer_struct

			default:

				wg.Done()
				return
			}
		}
	}()

	wg.Wait()

	Connector.RabbitMQ_channel.Close()
	Connector.InitRabbitMQ()

	return customer_map_json, nil

}

func (Connector *Connector) SendInQueue(Customer_struct rootsctuct.Customer_struct) error {

	if Connector.RabbitMQ_channel == nil {
		err := errors.New("Connection to RabbitMQ not established")
		return err
	}

	q, err := Connector.RabbitMQ_channel.QueueDeclare(
		"Customer___add_change", // name
		false,                   // durable
		false,                   // delete when unused
		false,                   // exclusive
		false,                   // no-wait
		nil,                     // arguments
	)
	if err != nil {
		return err
	}

	bodyJSON, err := json.Marshal(Customer_struct)
	if err != nil {
		return err
	}

	err = Connector.RabbitMQ_channel.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        bodyJSON,
		})

	if err != nil {
		return err
	}

	return nil

}

func (Connector *Connector) InitRabbitMQ() error {

	// Experimenting with RabbitMQ on your workstation? Try the community Docker image:
	// docker run -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management

	conn, err := amqp.Dial(Connector.Global_settings.AddressRabbitMQ) //5672
	if err != nil {
		Connector.LoggerCRM.ErrorLogger.Println("Failed to connect to RabbitMQ")
		Connector.RabbitMQ_channel = nil
		return err
	}
	//defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		Connector.LoggerCRM.ErrorLogger.Println("Failed to open a channel")
		Connector.RabbitMQ_channel = nil
		return err
	}
	//defer ch.Close()

	Connector.RabbitMQ_channel = ch

	return nil
}
