package rootdescription

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

var Global_settingsV Global_settings

var LoggerConnV LoggerConn

type Customer_struct struct {
	Customer_id    string
	Customer_name  string
	Customer_type  string
	Customer_email string
	Address_Struct Address_Struct
}

type Address_Struct struct {
	Street string
	House  int
}

type Global_settings struct {
	DataBaseType            string
	AddressRabbitMQ         string
	AddressRedis            string
	AddressMongoBD          string
	UseRabbitMQ             bool
	ElasticSearchAdress9200 string
	ElasticSearchAdress9300 string
	ElasticSearchIndexName  string
	Enterprise1CAdress      string
}

type Odata1C struct {
	Metadata string       `json:"odata.metadata"`
	Value    []Odata1Cdoc `json:"value"`
}

type Odata1Cdoc struct {
	Ref_Key            string
	DataVersion        string
	DeletionMark       bool
	Code               string
	Description        string
	Id                 string `json:"Идентификатор"`
	Name               string `json:"Имя"`
	Type               string `json:"Тип"`
	Email              string `json:"Майл"`
	Predefined         bool
	PredefinedDataName string
}

func (GlobalSettings *Global_settings) SaveSettingsOnDisk() {

	f, err := os.Create("./settings/config.json")
	if err != nil {
		log.Fatal(err)
	}

	JsonString, err := json.Marshal(GlobalSettings)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := f.Write(JsonString); err != nil {
		f.Close() // ignore error; Write error takes precedence
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func (GlobalSettings *Global_settings) LoadSettingsFromDisk() {

	file, err := os.OpenFile("./settings/config.json", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		file, err = ioutil.TempFile("", "ConfigJsonlogConnector1C")
		if err != nil {
			fmt.Println(err)
		}
	}

	decoder := json.NewDecoder(file)
	Settings := Global_settings{}
	err = decoder.Decode(&Settings)
	if err != nil {
		fmt.Println(err)
	}

	//Global_settingsV = Settings
	*GlobalSettings = Settings

	if err := file.Close(); err != nil {
		fmt.Println(err)
	}
}

type LoggerConn struct {
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
}

func (LoggerConn *LoggerConn) InitLog() {

	file, err := os.OpenFile("./logs/logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		file, err = ioutil.TempFile("", "logCoonector1C")
		if err != nil {
			log.Fatal(err)
		}
	}

	LoggerConn.InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	LoggerConn.ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	LoggerConn.ErrorLogger.Println("Starting the application...")
}

type ViewData struct {
	Title        string
	Message      string
	User         string
	DataBaseType string
	Customers    map[string]Customer_struct
}

type Log1C struct {
	Level                   string
	Date                    string
	ApplicationName         string
	ApplicationPresentation string
	Event                   string
	EventPresentation       string
	User                    string
	UserName                string
	Computer                string
	Metadata                string
	MetadataPresentation    string
	Comment                 string
	Data                    string
	DataPresentation        string
	TransactionStatus       string
	TransactionID           string
	Connection              string
	Session                 string
	ServerName              string
	Port                    string
	SyncPort                string
}

// Генератор структуры для XML https://www.onlinetool.io/xmltogo/
// type EventLog1C struct {
// 	XMLName xml.Name `xml:"EventLog"`
// 	Text    string   `xml:",chardata"`
// 	V8e     string   `xml:"v8e,attr"`
// 	Xs      string   `xml:"xs,attr"`
// 	Xsi     string   `xml:"xsi,attr"`
// 	Event   []struct {
// 		Text                    string `xml:",chardata"`
// 		Level                   string `xml:"Level"`
// 		Date                    string `xml:"Date"`
// 		ApplicationName         string `xml:"ApplicationName"`
// 		ApplicationPresentation string `xml:"ApplicationPresentation"`
// 		Event                   string `xml:"Event"`
// 		EventPresentation       string `xml:"EventPresentation"`
// 		User                    string `xml:"User"`
// 		UserName                string `xml:"UserName"`
// 		Computer                string `xml:"Computer"`
// 		Metadata                string `xml:"Metadata"`
// 		MetadataPresentation    string `xml:"MetadataPresentation"`
// 		Comment                 string `xml:"Comment"`
// 		Data                    struct {
// 			Text          string `xml:",chardata"`
// 			Nil           string `xml:"nil,attr"`
// 			Name          string `xml:"Name"`
// 			CurrentOSUser string `xml:"CurrentOSUser"`
// 		} `xml:"Data"`
// 		DataPresentation  string `xml:"DataPresentation"`
// 		TransactionStatus string `xml:"TransactionStatus"`
// 		TransactionID     string `xml:"TransactionID"`
// 		Connection        string `xml:"Connection"`
// 		Session           string `xml:"Session"`
// 		ServerName        string `xml:"ServerName"`
// 		Port              string `xml:"Port"`
// 		SyncPort          string `xml:"SyncPort"`
// 	} `xml:"Event"`
// }

// Генератор структуры для XML https://www.onlinetool.io/xmltogo/
type EventLog1C struct {
	XMLName xml.Name  `xml:"EventLog"`
	Text    string    `xml:",chardata"`
	V8e     string    `xml:"v8e,attr"`
	Xs      string    `xml:"xs,attr"`
	Xsi     string    `xml:"xsi,attr"`
	Event   []Event1C `xml:"Event"`
}

type Event1C struct {
	Text                    string `xml:",chardata"`
	Level                   string `xml:"Level"`
	Date                    string `xml:"Date"`
	ApplicationName         string `xml:"ApplicationName"`
	ApplicationPresentation string `xml:"ApplicationPresentation"`
	Event                   string `xml:"Event"`
	EventPresentation       string `xml:"EventPresentation"`
	User                    string `xml:"User"`
	UserName                string `xml:"UserName"`
	Computer                string `xml:"Computer"`
	Metadata                string `xml:"Metadata"`
	MetadataPresentation    string `xml:"MetadataPresentation"`
	Comment                 string `xml:"Comment"`
	// Data                    struct {
	// 	Text          string `xml:",chardata"`
	// 	Nil           string `xml:"nil,attr"`
	// 	Name          string `xml:"Name"`
	// 	CurrentOSUser string `xml:"CurrentOSUser"`
	// } `xml:"Data"`
	Data              string `xml:"Data"`
	DataPresentation  string `xml:"DataPresentation"`
	TransactionStatus string `xml:"TransactionStatus"`
	TransactionID     string `xml:"TransactionID"`
	Connection        string `xml:"Connection"`
	Session           string `xml:"Session"`
	ServerName        string `xml:"ServerName"`
	Port              string `xml:"Port"`
	SyncPort          string `xml:"SyncPort"`
}

type Event1CExtended struct {
	Text                    string              `xml:",chardata"`
	Level                   string              `xml:"Level"`
	Date                    string              `xml:"Date"`
	ApplicationName         string              `xml:"ApplicationName"`
	ApplicationPresentation string              `xml:"ApplicationPresentation"`
	Event                   string              `xml:"Event"`
	EventPresentation       string              `xml:"EventPresentation"`
	User                    string              `xml:"User"`
	UserName                string              `xml:"UserName"`
	Computer                string              `xml:"Computer"`
	Metadata                string              `xml:"Metadata"`
	MetadataPresentation    string              `xml:"MetadataPresentation"`
	Comment                 string              `xml:"Comment"`
	Data                    Event1CExtendedData `xml:"Data"`
	DataPresentation        string              `xml:"DataPresentation"`
	TransactionStatus       string              `xml:"TransactionStatus"`
	TransactionID           string              `xml:"TransactionID"`
	Connection              string              `xml:"Connection"`
	Session                 string              `xml:"Session"`
	ServerName              string              `xml:"ServerName"`
	Port                    string              `xml:"Port"`
	SyncPort                string              `xml:"SyncPort"`
}

type EventLog1CExtended struct {
	XMLName xml.Name          `xml:"EventLog"`
	Text    string            `xml:",chardata"`
	V8e     string            `xml:"v8e,attr"`
	Xs      string            `xml:"xs,attr"`
	Xsi     string            `xml:"xsi,attr"`
	Event   []Event1CExtended `xml:"Event"`
}

type Event1CExtendedData struct {
	Text          string
	Nil           string
	Name          string
	CurrentOSUser string
}
