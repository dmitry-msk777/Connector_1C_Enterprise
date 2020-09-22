package rootdescription

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

var Global_settingsV Global_settings

var LoggerCRMv LoggerCRM

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
	DataBaseType    string
	AddressRabbitMQ string
	UseRabbitMQ     bool
}

func (GlobalSettings *Global_settings) SaveSettingsOnDisk() {

	f, err := os.Create("./settings/config.json")
	if err != nil {
		log.Fatal(err)
	}

	JsonString, err := json.Marshal(GlobalSettings)
	if err != nil {
		//EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
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

type LoggerCRM struct {
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
}

func (LoggerCRM *LoggerCRM) InitLog() {

	file, err := os.OpenFile("./logs/logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		file, err = ioutil.TempFile("", "logCoonector1C")
		if err != nil {
			log.Fatal(err)
		}
	}

	LoggerCRM.InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	LoggerCRM.ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	LoggerCRM.ErrorLogger.Println("Starting the application...")
}

type ViewData struct {
	Title        string
	Message      string
	User         string
	DataBaseType string
	Customers    map[string]Customer_struct
}
