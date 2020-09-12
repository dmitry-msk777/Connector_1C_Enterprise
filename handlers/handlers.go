package handlers

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"

	connector "github.com/dmitry-msk777/Connector_1C_Enterprise/connector"
	rootsctuct "github.com/dmitry-msk777/Connector_1C_Enterprise/rootdescription"

	"encoding/json"
	"io/ioutil"
)

func Settings(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {

		tmpl, err := template.ParseFiles("templates/settings.html", "templates/header.html")
		if err != nil {
			connector.ConnectorV.LoggerCRM.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
			return
		}

		tmpl.ExecuteTemplate(w, "settings", connector.ConnectorV.Global_settings)

	} else {

		rootsctuct.Global_settingsV.AddressRabbitMQ = r.FormValue("AddressRabbitMQ")
		rootsctuct.Global_settingsV.DataBaseType = r.FormValue("DataBaseType")

		if r.FormValue("UseRabbitMQ") == "on" {
			rootsctuct.Global_settingsV.UseRabbitMQ = true
		} else {
			rootsctuct.Global_settingsV.UseRabbitMQ = false
		}

		rootsctuct.Global_settingsV.ElasticSearchAdress9200 = r.FormValue("ElasticSearchAdress9200")
		rootsctuct.Global_settingsV.ElasticSearchAdress9300 = r.FormValue("ElasticSearchAdress9300")
		rootsctuct.Global_settingsV.ElasticSearchIndexName = r.FormValue("ElasticSearchIndexName")

		connector.ConnectorV.SetSettings(rootsctuct.Global_settingsV)

		err := connector.ConnectorV.InitDataBase()
		if err != nil {
			connector.ConnectorV.LoggerCRM.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
			return
		}

		connector.ConnectorV.Global_settings.SaveSettingsOnDisk()

		http.Redirect(w, r, "/", 302)
	}
}

func RabbitMQ_1C(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {

		customer_map_json, err := connector.ConnectorV.ConsumeFromQueue()

		if err != nil {
			connector.ConnectorV.LoggerCRM.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
			return
		}

		JsonString, err := json.Marshal(customer_map_json)
		if err != nil {
			connector.ConnectorV.LoggerCRM.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, "error json:"+err.Error())
			return
		}
		fmt.Fprintf(w, string(JsonString))

	} else {

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			connector.ConnectorV.LoggerCRM.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		var customer_map_json = make(map[string]rootsctuct.Customer_struct)

		err = json.Unmarshal(body, &customer_map_json)
		if err != nil {
			connector.ConnectorV.LoggerCRM.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		for _, p := range customer_map_json {

			if connector.ConnectorV.Global_settings.UseRabbitMQ {
				err = connector.ConnectorV.SendInQueue(p)
				if err != nil {
					connector.ConnectorV.LoggerCRM.ErrorLogger.Println(err.Error())
					fmt.Fprintf(w, err.Error())
					return
				}
			}

		}

		fmt.Fprintf(w, string(body))

	}

}

func log1C_xml(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {

		// Пока по GET ничего не делаем.

		// customer_map_s, err := enginecrm.EngineCRMv.GetAllCustomer(enginecrm.EngineCRMv.DataBaseType)

		// if err != nil {
		// 	enginecrm.EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
		// 	fmt.Fprintf(w, err.Error())
		// 	return
		// }

		// doc := etree.NewDocument()
		// //doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)

		// Custromers := doc.CreateElement("Custromers")

		// for _, p := range customer_map_s {
		// 	Custromer := Custromers.CreateElement("Custromer")
		// 	Custromer.CreateAttr("value", p.Customer_id)

		// 	id := Custromer.CreateElement("Customer_id")
		// 	id.CreateAttr("value", p.Customer_id)
		// 	name := Custromer.CreateElement("Customer_name")
		// 	name.CreateAttr("value", p.Customer_name)
		// 	type1 := Custromer.CreateElement("Customer_type")
		// 	type1.CreateAttr("value", p.Customer_type)
		// 	email := Custromer.CreateElement("Customer_email")
		// 	email.CreateAttr("value", p.Customer_email)
		// }

		// //doc.CreateText("/xml")

		// doc.Indent(2)
		// XMLString, _ := doc.WriteToString()

		// fmt.Fprintf(w, XMLString)

	} else {

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			connector.ConnectorV.LoggerCRM.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		Log1C_slice, err := connector.ConnectorV.ParseXMLFrom1C(body)
		if err != nil {
			connector.ConnectorV.LoggerCRM.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		err = connector.ConnectorV.SendInElastichSearch(Log1C_slice)

		if err != nil {
			connector.ConnectorV.LoggerCRM.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		fmt.Fprintf(w, "Succeed!")

	}
}

func StratHandlers() {

	router := mux.NewRouter()

	router.HandleFunc("/", Settings)
	router.HandleFunc("/settings", Settings)

	router.HandleFunc("/rabbitMQ_1C", RabbitMQ_1C)
	router.HandleFunc("/log1C_xml", log1C_xml)

	http.Handle("/", router)
	http.ListenAndServe(":8181", nil)
	fmt.Println("Server is listening...")

}
