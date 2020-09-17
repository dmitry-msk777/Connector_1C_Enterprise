package connector

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/streadway/amqp"

	"encoding/json"

	"github.com/beevik/etree"
	"github.com/olivere/elastic"

	"github.com/go-redis/redis/v7"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	rootsctuct "github.com/dmitry-msk777/Connector_1C_Enterprise/rootdescription"
)

var ConnectorV Connector

type Connector struct {
	DataBaseType      string
	RabbitMQ_channel  *amqp.Channel
	Global_settings   rootsctuct.Global_settings
	LoggerCRM         rootsctuct.LoggerCRM
	CollectionMongoDB *mongo.Collection
	DemoDBmap         map[string]rootsctuct.Customer_struct
	RedisClient       *redis.Client
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

	switch Connector.DataBaseType {
	case "Redis":
		//localhost:32769
		Connector.RedisClient = intiRedisClient(Connector.Global_settings.AddressRedis)

		pong, err := Connector.RedisClient.Ping().Result()
		if err != nil {
			Connector.RedisClient = nil
			fmt.Println(pong, err)
			return err
		}

	case "MongoDB":

		//temporary
		//collectionMongoDB = GetCollectionMongoBD("CRM", "customers", "mongodb://localhost:32768")
		//"mongodb://localhost:32768"
		Connector.CollectionMongoDB = GetCollectionMongoBD("CRM", "customers", Connector.Global_settings.AddressMongoBD)

	default:

		var ArrayCustomer []rootsctuct.Customer_struct

		ArrayCustomer = append(ArrayCustomer, rootsctuct.Customer_struct{
			Customer_id:    "777",
			Customer_name:  "Dmitry",
			Customer_type:  "Cust",
			Customer_email: "fff@mail.ru",
		})

		ArrayCustomer = append(ArrayCustomer, rootsctuct.Customer_struct{
			Customer_id:    "666",
			Customer_name:  "Alex",
			Customer_type:  "Cust_Fiz",
			Customer_email: "44fish@mail.ru",
		})

		var mapForEngineCRM = make(map[string]rootsctuct.Customer_struct)
		Connector.DemoDBmap = mapForEngineCRM

		for _, p := range ArrayCustomer {
			Connector.DemoDBmap[p.Customer_id] = p
		}

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

func (Connector *Connector) ParseXMLFrom1C(body []byte) ([]rootsctuct.Log1C, error) {
	doc := etree.NewDocument()

	if err := doc.ReadFromBytes(body); err != nil {
		return nil, err
	}

	// var customer_map_xml = make(map[string]rootsctuct.Customer_struct)
	var Log1C_slice []rootsctuct.Log1C

	// Custromers := doc.SelectElement("Custromers")
	EventLog := doc.SelectElement("v8e:EventLog")

	// for _, Custromer := range Custromers.SelectElements("Custromer") {

	for _, Event := range EventLog.SelectElements("v8e:Event") {

		// 	Customer_struct := rootsctuct.Customer_struct{}
		// 	//fmt.Println("CHILD element:", Custromer.Tag)
		Log1C := rootsctuct.Log1C{}

		if v8e_Level := Event.SelectElement("v8e:Level"); v8e_Level != nil {
			//value := v8e_Level.SelectAttrValue("value", "unknown")
			Log1C.Level = v8e_Level.Text()
			//Log1C.Level = v8e_Level.Child[0].Data
		}

		if v8e_Date := Event.SelectElement("v8e:Date"); v8e_Date != nil {
			Log1C.Date = v8e_Date.Text()
		}

		if v8e_ApplicationName := Event.SelectElement("v8e:ApplicationName"); v8e_ApplicationName != nil {
			Log1C.ApplicationName = v8e_ApplicationName.Text()
		}

		if v8e_ApplicationPresentation := Event.SelectElement("v8e:ApplicationPresentation"); v8e_ApplicationPresentation != nil {
			Log1C.ApplicationPresentation = v8e_ApplicationPresentation.Text()
		}

		if v8e_Event := Event.SelectElement("v8e:Event"); v8e_Event != nil {
			Log1C.Event = v8e_Event.Text()
		}

		if v8e_EventPresentation := Event.SelectElement("v8e:EventPresentation"); v8e_EventPresentation != nil {
			Log1C.EventPresentation = v8e_EventPresentation.Text()
		}

		if v8e_User := Event.SelectElement("v8e:User"); v8e_User != nil {
			Log1C.User = v8e_User.Text()
		}

		if v8e_UserName := Event.SelectElement("v8e:UserName"); v8e_UserName != nil {
			Log1C.UserName = v8e_UserName.Text()
		}

		if v8e_Computer := Event.SelectElement("v8e:Computer"); v8e_Computer != nil {
			Log1C.Computer = v8e_Computer.Text()
		}

		if v8e_Metadata := Event.SelectElement("v8e:Metadata"); v8e_Metadata != nil {
			Log1C.Metadata = v8e_Metadata.Text()
		}

		if v8e_MetadataPresentation := Event.SelectElement("v8e:MetadataPresentation"); v8e_MetadataPresentation != nil {
			Log1C.MetadataPresentation = v8e_MetadataPresentation.Text()
		}

		if v8e_Comment := Event.SelectElement("v8e:Comment"); v8e_Comment != nil {
			Log1C.Comment = v8e_Comment.Text()
		}

		if v8e_Data := Event.SelectElement("v8e:Data"); v8e_Data != nil {
			Log1C.Data = v8e_Data.Text()
		}

		if v8e_DataPresentation := Event.SelectElement("v8e:DataPresentation"); v8e_DataPresentation != nil {
			Log1C.DataPresentation = v8e_DataPresentation.Text()
		}

		if v8e_TransactionStatus := Event.SelectElement("v8e:TransactionStatus"); v8e_TransactionStatus != nil {
			Log1C.TransactionStatus = v8e_TransactionStatus.Text()
		}

		if v8e_TransactionID := Event.SelectElement("v8e:TransactionID"); v8e_TransactionID != nil {
			Log1C.TransactionID = v8e_TransactionID.Text()
		}

		if v8e_Connection := Event.SelectElement("v8e:Connection"); v8e_Connection != nil {
			Log1C.Connection = v8e_Connection.Text()
		}

		if v8e_Session := Event.SelectElement("v8e:Session"); v8e_Session != nil {
			Log1C.Session = v8e_Session.Text()
		}

		if v8e_ServerName := Event.SelectElement("v8e:ServerName"); v8e_ServerName != nil {
			Log1C.ServerName = v8e_ServerName.Text()
		}

		if v8e_Port := Event.SelectElement("v8e:Port"); v8e_Port != nil {
			Log1C.Port = v8e_Port.Text()
		}

		if v8e_SyncPort := Event.SelectElement("v8e:SyncPort"); v8e_SyncPort != nil {
			Log1C.SyncPort = v8e_SyncPort.Text()
		}

		Log1C_slice = append(Log1C_slice, Log1C)
	}

	//fmt.Println(Log1C_slice)

	return Log1C_slice, nil
}

func (Connector *Connector) SendInElastichSearch(Log1C_slice []rootsctuct.Log1C) error {

	// clientElasticSerch, err := elastic.NewClient(elastic.SetSniff(false),
	// 	elastic.SetURL("http://127.0.0.1:9200", "http://127.0.0.1:9300"))
	//// elastic.SetBasicAuth("user", "secret"))

	clientElasticSerch, err := elastic.NewClient(elastic.SetSniff(false),
		elastic.SetURL(Connector.Global_settings.ElasticSearchAdress9200, Connector.Global_settings.ElasticSearchAdress9300))

	if err != nil {
		return err
	}

	// index example "transactionid"
	exists, err := clientElasticSerch.IndexExists(Connector.Global_settings.ElasticSearchIndexName).Do(context.Background())
	if err != nil {
		return err
	}

	if !exists {
		// Create a new index.
		mapping := `
				{
					"settings":{
						"number_of_shards":1,
						"number_of_replicas":0
					},
					"mappings":{
						"doc":{
							"properties":{
								"Level":{
									"type":"text"
								},
								"Date":{
									"type":"text"
								},
								"ApplicationName":{
									"type":"text"
								},
								"ApplicationPresentation":{
									"type":"text"
								},
								"Event":{
									"type":"text"
								},
								"EventPresentation":{
									"type":"text"
								},
								"User":{
									"type":"text"
								},
								"UserName":{
									"type":"text"
								},
								"Computer":{
									"type":"text"
								},
								"Metadata":{
									"type":"text"
								},
								"MetadataPresentation":{
									"type":"text"
								},
								"Comment":{
									"type":"text"
								},
								"Data":{
									"type":"text"
								},
								"DataPresentation":{
									"type":"text"
								},
								"TransactionStatus":{
									"type":"text"
								},
								"TransactionID":{
									"type":"text",
									"store": true,
									"fielddata": true
								},
								"Connection":{
									"type":"text"
								},
								"Session":{
									"type":"text"
								},
								"ServerName":{
									"type":"text"
								},
								"Port":{
									"type":"text"
								},
								"SyncPort":{
									"type":"text"
								}
						}
					}
				}
				}`

		//createIndex, err := clientElasticSerch.CreateIndex("TransactionID").Body(mapping).IncludeTypeName(true).Do(context.Background())
		createIndex, err := clientElasticSerch.CreateIndex(Connector.Global_settings.ElasticSearchIndexName).Body(mapping).Do(context.Background())
		if err != nil {
			return err
		}
		if !createIndex.Acknowledged {
		}
	}

	for _, p := range Log1C_slice {

		put1, err := clientElasticSerch.Index().
			Index("transactionid").
			Type("doc").
			Id(p.TransactionID).
			BodyJson(p).
			Do(context.Background())
		if err != nil {
			Connector.LoggerCRM.ErrorLogger.Println(err.Error())
			//fmt.Fprintf(w, err.Error())
			return err
		}
		fmt.Printf("Indexed record %s to index %s, type %s\n", put1.Id, put1.Index, put1.Type)

	}

	// Flush to make sure the documents got written.
	_, err = clientElasticSerch.Flush().Index(Connector.Global_settings.ElasticSearchIndexName).Do(context.Background())
	if err != nil {
		return err
	}

	// // +++ Search with a term query
	// termQuery := elastic.NewTermQuery("TransactionID", "11.09.2020 15:12:07 (1446734)")
	// searchResult, err := clientElasticSerch.Search().
	// 	Index("transactionid").      // search in index "crm_customer"
	// 	Query(termQuery).            // specify the query
	// 	Sort("TransactionID", true). // sort by "user" field, ascending
	// 	From(0).Size(10).            // take documents 0-9
	// 	Pretty(true).                // pretty print request and response JSON
	// 	Do(context.Background())     // execute
	// if err != nil {
	// 	return err
	// }

	// // +++ searchResult is of type SearchResult and returns hits, suggestions,
	// // and all kinds of other information from Elasticsearch.
	// fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)

	// var ttyp rootsctuct.Log1C
	// for _, item := range searchResult.Each(reflect.TypeOf(ttyp)) {
	// 	t := item.(rootsctuct.Log1C)
	// 	//fmt.Fprintf(w, "customer_id: %s customer_name: %s", t.TransactionID, t.TransactionID)
	// 	fmt.Printf("TransactionID: %s", t.TransactionID)
	// }
	// fmt.Printf("Found a total of %d records\n", searchResult.TotalHits())

	// // // +++ Delete an index.
	// // deleteIndex, err := clientElasticSerch.DeleteIndex("transactionid").Do(context.Background())
	// // if err != nil {
	// // 	return err
	// // }
	// // if !deleteIndex.Acknowledged {
	// // 	// Not acknowledged
	// // }

	return nil

}

func (Connector *Connector) GetAllCustomer(DataBaseType string) (map[string]rootsctuct.Customer_struct, error) {

	var customer_map_s = make(map[string]rootsctuct.Customer_struct)

	switch DataBaseType {
	case "MongoDB":

		cur, err := Connector.CollectionMongoDB.Find(context.Background(), bson.D{})
		if err != nil {
			return customer_map_s, err
		}
		defer cur.Close(context.Background())

		Customer_struct_slice := []rootsctuct.Customer_struct{}

		for cur.Next(context.Background()) {

			Customer_struct_out := rootsctuct.Customer_struct{}

			err := cur.Decode(&Customer_struct_out)
			if err != nil {
				return customer_map_s, err
			}

			Customer_struct_slice = append(Customer_struct_slice, Customer_struct_out)

			// To get the raw bson bytes use cursor.Current
			// // raw := cur.Current
			// // fmt.Println(raw)
			// do something with raw...
		}
		if err := cur.Err(); err != nil {
			return customer_map_s, err
		}

		for _, p := range Customer_struct_slice {
			customer_map_s[p.Customer_id] = p
		}

		return customer_map_s, nil

	case "Redis":

		var cursor uint64
		ScanCmd := Connector.RedisClient.Scan(cursor, "", 100)
		//fmt.Println(ScanCmd)

		cursor1, _, err := ScanCmd.Result()

		if err != nil {
			Connector.LoggerCRM.ErrorLogger.Println("key2 does not exist")
			return customer_map_s, err
		}

		//fmt.Println(cursor1, keys1)

		Customer_struct_slice := []rootsctuct.Customer_struct{}
		for _, value := range cursor1 {
			p := rootsctuct.Customer_struct{}
			//IDString := strconv.FormatInt(int64(i), 10)
			val2, err := Connector.RedisClient.Get(value).Result()
			if err == redis.Nil {
				Connector.LoggerCRM.ErrorLogger.Println("key2 does not exist")
				continue
				//fmt.Println("key2 does not exist")
			} else if err != nil {
				Connector.LoggerCRM.ErrorLogger.Println(err.Error())
				continue
			} else {
				//fmt.Println("key2", val2)

				err = json.Unmarshal([]byte(val2), &p)
				if err != nil {
					Connector.LoggerCRM.ErrorLogger.Println(err.Error())
					continue
				}

				Customer_struct_slice = append(Customer_struct_slice, p)
			}
		}

		for _, p := range Customer_struct_slice {
			customer_map_s[p.Customer_id] = p
		}

		return customer_map_s, nil

	default:
		return Connector.DemoDBmap, nil
	}

}

func (Connector *Connector) AddChangeOneRow(DataBaseType string, Customer_struct rootsctuct.Customer_struct, Global_settings rootsctuct.Global_settings) error {

	switch DataBaseType {

	case "MongoDB":

		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

		//maybe use? insertMany(): добавляет несколько документов
		//before adding find db.users.find()

		insertResult, err := Connector.CollectionMongoDB.InsertOne(ctx, Customer_struct)
		if err != nil {
			return err
		}
		fmt.Println(insertResult.InsertedID)

		// This update function can use the separate Update and Paste pre-search?

		// opts := options.Update().SetUpsert(true)
		// filter := bson.D{{"customer_id", Customer_struct.Customer_id}}
		// update := bson.D{{"$set", bson.D{{"customer_name", Customer_struct.Customer_name}, {"customer_type", Customer_struct.Customer_type}, {"customer_email", Customer_struct.Customer_email}}}}

		// result, err := EngineCRMv.collectionMongoDB.UpdateOne(context.TODO(), filter, update, opts)
		// if err != nil {
		// 	ErrorLogger.Println(err.Error())
		// 	return err.Error()
		// }

		// if result.MatchedCount != 0 {
		// 	fmt.Println("matched and replaced an existing document")
		// }
		// if result.UpsertedCount != 0 {
		// 	fmt.Printf("inserted a new document with ID %v\n", result.UpsertedID)
		// }
	case "Redis":

		JsonStr, err := json.Marshal(Customer_struct)
		if err != nil {
			return err
		}

		err = Connector.RedisClient.Set(Customer_struct.Customer_id, string(JsonStr), 0).Err()
		if err != nil {
			return err
		}

	default:
		Connector.DemoDBmap[Customer_struct.Customer_id] = Customer_struct
	}

	return nil
}

func intiRedisClient(Addr string) *redis.Client {

	client := redis.NewClient(&redis.Options{
		Addr:     Addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return client
}

func GetCollectionMongoBD(Database string, Collection string, HostConnect string) *mongo.Collection {

	clientOptions := options.Client().ApplyURI(HostConnect)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		fmt.Println(err.Error())
	}
	err = client.Connect(context.Background())
	if err != nil {
		fmt.Println(err.Error())
	}

	err = client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		fmt.Println("Couldn't connect to the database", err.Error())
	} else {
		fmt.Println("Connected MongoDB!")
	}

	return client.Database(Database).Collection(Collection)
}

func (Connector *Connector) FindOneRow(DataBaseType string, id string, Global_settings rootsctuct.Global_settings) (rootsctuct.Customer_struct, error) {

	Customer_struct_out := rootsctuct.Customer_struct{}

	switch DataBaseType {
	case "MongoDB":

		err := Connector.CollectionMongoDB.FindOne(context.TODO(), bson.D{{"customer_id", id}}).Decode(&Customer_struct_out)
		if err != nil {
			// ErrNoDocuments means that the filter did not match any documents in the collection
			if err == mongo.ErrNoDocuments {
				return Customer_struct_out, err
			}
		}
		fmt.Printf("found document %v", Customer_struct_out)

	case "Redis":

		val2, err := Connector.RedisClient.Get(id).Result()
		if err == redis.Nil {
			Connector.LoggerCRM.ErrorLogger.Println("key2 does not exist")
			return Customer_struct_out, err
		} else if err != nil {
			Connector.LoggerCRM.ErrorLogger.Println(err.Error())
			return Customer_struct_out, err
		} else {
			err = json.Unmarshal([]byte(val2), &Customer_struct_out)
			if err != nil {
				Connector.LoggerCRM.ErrorLogger.Println(err.Error())
				return Customer_struct_out, err
			}

			return Customer_struct_out, nil
		}

	default:
		Customer_struct_out = Connector.DemoDBmap[id]
	}

	return Customer_struct_out, nil
}

func (Connector *Connector) DeleteOneRow(DataBaseType string, id string, Global_settings rootsctuct.Global_settings) error {

	switch DataBaseType {
	case "MongoDB":

		res, err := Connector.CollectionMongoDB.DeleteOne(context.TODO(), bson.D{{"customer_id", id}})
		if err != nil {
			return err
		}
		fmt.Printf("deleted %v documents\n", res.DeletedCount)

	case "Redis":

		//iter := EngineCRMv.RedisClient.Scan(0, "prefix*", 0).Iterator()
		iter := Connector.RedisClient.Scan(0, id, 0).Iterator()
		for iter.Next() {
			err := Connector.RedisClient.Del(iter.Val()).Err()
			if err != nil {
				Connector.LoggerCRM.ErrorLogger.Println(err.Error())
				return err
			}
			//fmt.Println(iter.Val())
		}
		if err := iter.Err(); err != nil {
			Connector.LoggerCRM.ErrorLogger.Println(err.Error())
			return err
		}

	default:
		_, ok := Connector.DemoDBmap[id]
		if ok {
			delete(Connector.DemoDBmap, id)
		}
	}

	return nil

}
