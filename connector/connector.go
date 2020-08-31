package connector

import (
	"errors"
	"fmt"
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

func (Connector *Connector) SetSettings(Global_settings rootsctuct.Global_settings) {

	Connector.DataBaseType = Global_settings.DataBaseType
	if Global_settings.DataBaseType == "" {
		Connector.DataBaseType = "DemoRegime"
		Global_settings.DataBaseType = "DemoRegime"
	}

	Connector.Global_settings = Global_settings

}

func (Connector *Connector) InitDataBase() error {

	if Connector.Global_settings.UseRabbitMQ && Connector.RabbitMQ_channel == nil {
		Connector.InitRabbitMQ(rootsctuct.Global_settingsV)
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

	// msgs, err := EngineCRM.RabbitMQ_channel.Consume(
	// 	q.Name, // queue
	// 	"",     // consumer
	// 	true,   // auto-ack
	// 	false,  // exclusive
	// 	false,  // no-local
	// 	false,  // no-wait
	// 	nil,    // args
	// )

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
	//////////////////////////////////////////////////////////////////////////
	//forever := make(chan bool)

	// go func() {

	// 	for d := range msgs {

	// 		Customer_struct := rootsctuct.Customer_struct{}

	// 		err = json.Unmarshal(d.Body, &Customer_struct)
	// 		if err != nil {
	// 			EngineCRM.LoggerCRM.ErrorLogger.Println(err.Error())
	// 		}
	// 		customer_map_json[Customer_struct.Customer_id] = Customer_struct
	// 		fmt.Println("Customer_struct.Customer_id = ", Customer_struct.Customer_id)
	// 	}

	// }()

	//signals := make(chan bool)
	///////////////////////////////////////////////////////////
	go func() {
		for {
			select {
			case msg := <-msgs:
				fmt.Println("msg = ", msg)
				Customer_struct := rootsctuct.Customer_struct{}

				err = json.Unmarshal(msg.Body, &Customer_struct)
				if err != nil {
					Connector.LoggerCRM.ErrorLogger.Println(err.Error())
				}

				customer_map_json[Customer_struct.Customer_id] = Customer_struct

				//msg.Ack(false)

				// if Customer_struct.Customer_id != "" {
				// 	customer_map_json[Customer_struct.Customer_id] = Customer_struct
				// }

				//fmt.Println("Customer_struct.Customer_id = ", Customer_struct.Customer_id)

			// case <-signals:
			// 	return
			default:
				//fmt.Println("<-- loop broke!")
				return // exit break loop
			}
		}
	}()

	/////////////////////////////////////////////////////////////////////

	time.Sleep(10000 * time.Millisecond)
	Connector.RabbitMQ_channel.Close()
	//Connector.RabbitMQ_channel = nil
	Connector.InitRabbitMQ(rootsctuct.Global_settingsV)

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

func (Connector *Connector) InitRabbitMQ(Global_settings rootsctuct.Global_settings) error {

	// Experimenting with RabbitMQ on your workstation? Try the community Docker image:
	// docker run -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management

	conn, err := amqp.Dial(Global_settings.AddressRabbitMQ) //5672
	if err != nil {
		Connector.LoggerCRM.ErrorLogger.Println("Failed to connect to RabbitMQ")
		return err
	}
	//defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		Connector.LoggerCRM.ErrorLogger.Println("Failed to open a channel")
		return err
	}
	//defer ch.Close()

	Connector.RabbitMQ_channel = ch

	return nil
}
