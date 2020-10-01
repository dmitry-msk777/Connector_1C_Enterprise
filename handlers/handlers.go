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

	"github.com/beevik/etree"
)

func Settings(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {

		tmpl, err := template.ParseFiles("templates/settings.html", "templates/header.html")
		if err != nil {
			connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
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

		rootsctuct.Global_settingsV.AddressRedis = r.FormValue("AddressRedis")
		rootsctuct.Global_settingsV.AddressMongoBD = r.FormValue("AddressMongoBD")

		rootsctuct.Global_settingsV.Enterprise1CAdress = r.FormValue("Enterprise1CAdress")

		connector.ConnectorV.SetSettings(rootsctuct.Global_settingsV)

		err := connector.ConnectorV.InitDataBase()
		if err != nil {
			connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
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
			connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
			return
		}

		JsonString, err := json.Marshal(customer_map_json)
		if err != nil {
			connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, "error json:"+err.Error())
			return
		}
		fmt.Fprintf(w, string(JsonString))

	} else {

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		var customer_map_json = make(map[string]rootsctuct.Customer_struct)

		err = json.Unmarshal(body, &customer_map_json)
		if err != nil {
			connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		for _, p := range customer_map_json {

			if connector.ConnectorV.Global_settings.UseRabbitMQ {
				err = connector.ConnectorV.SendInQueue(p)
				if err != nil {
					connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
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
			connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		Log1C_slice, err := connector.ConnectorV.ParseXMLFrom1C(body)
		if err != nil {
			connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		err = connector.ConnectorV.SendInElastichSearch(Log1C_slice)

		if err != nil {
			connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		fmt.Fprintf(w, "Succeed!")

	}
}

func Api_json(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":

		customer_map_s, err := connector.ConnectorV.GetAllCustomer(connector.ConnectorV.DataBaseType)

		if err != nil {
			connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
			return
		}

		JsonString, err := json.Marshal(customer_map_s)
		if err != nil {
			connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, "error json:"+err.Error())
		}
		fmt.Fprintf(w, string(JsonString))

	case "POST":

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		var customer_map_json = make(map[string]rootsctuct.Customer_struct)

		err = json.Unmarshal(body, &customer_map_json)
		if err != nil {
			connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		for _, p := range customer_map_json {
			err := connector.ConnectorV.AddChangeOneRow(connector.ConnectorV.DataBaseType, p, rootsctuct.Global_settingsV)
			if err != nil {
				connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
				fmt.Println(err.Error())
			}
		}

		fmt.Fprintf(w, string(body))

	case "PUT":

		fmt.Fprintf(w, "PUT")

	case "DELETE":

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		var customer_map_json = make(map[string]rootsctuct.Customer_struct)

		err = json.Unmarshal(body, &customer_map_json)
		if err != nil {
			connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		for _, p := range customer_map_json {
			err := connector.ConnectorV.DeleteOneRow(connector.ConnectorV.DataBaseType, p.Customer_id, rootsctuct.Global_settingsV)
			if err != nil {
				connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
				fmt.Println(err.Error())
			}
		}

		fmt.Fprintf(w, string(body))

	default:

		fmt.Fprintf(w, r.Method+" - This method is not implemented")

	}

}

func List_customer(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("templates/list_customer.html", "templates/header.html")
	if err != nil {
		connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
		return
	}

	customer_map_data, err := connector.ConnectorV.GetAllCustomer(connector.ConnectorV.DataBaseType)

	if err != nil {
		connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
		return
	}

	tmpl.ExecuteTemplate(w, "list_customer", customer_map_data)

}

func EditPage(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["id"]

	Customer_struct_out, err := connector.ConnectorV.FindOneRow(connector.ConnectorV.DataBaseType, id, rootsctuct.Global_settingsV)

	if err != nil {
		connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
		return
	}

	tmpl, err := template.ParseFiles("templates/edit.html", "templates/header.html")
	if err != nil {
		connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
		return
	}

	tmpl.ExecuteTemplate(w, "edit", Customer_struct_out)

}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["id"]

	err := connector.ConnectorV.DeleteOneRow(connector.ConnectorV.DataBaseType, id, rootsctuct.Global_settingsV)

	if err != nil {
		connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
		return
	}

	http.Redirect(w, r, "/list_customer", 301)

}

func EditHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
	}

	Customer_struct_out := rootsctuct.Customer_struct{
		Customer_id:    r.FormValue("customer_id"),
		Customer_name:  r.FormValue("customer_name"),
		Customer_type:  r.FormValue("customer_type"),
		Customer_email: r.FormValue("customer_email"),
	}

	connector.ConnectorV.AddChangeOneRow(connector.ConnectorV.DataBaseType, Customer_struct_out, rootsctuct.Global_settingsV)

	//return err
	//fmt.Fprintf(w, err.Error())

	http.Redirect(w, r, "/list_customer", 301)

}

func Add_change_customer(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("templates/add_change_customer.html", "templates/header.html")
	if err != nil {
		connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
		return
	}

	tmpl.ExecuteTemplate(w, "add_change_customer", nil)

}

func Postform_add_change_customer(w http.ResponseWriter, r *http.Request) {

	customer_data := rootsctuct.Customer_struct{
		Customer_name:  r.FormValue("customer_name"),
		Customer_id:    r.FormValue("customer_id"),
		Customer_type:  r.FormValue("customer_type"),
		Customer_email: r.FormValue("customer_email"),
	}

	err := connector.ConnectorV.AddChangeOneRow(connector.ConnectorV.DataBaseType, customer_data, rootsctuct.Global_settingsV)

	if err != nil {
		connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
		return
	}

	http.Redirect(w, r, "/list_customer", 302)
}

func Api_xml(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":

		customer_map_s, err := connector.ConnectorV.GetAllCustomer(connector.ConnectorV.DataBaseType)

		if err != nil {
			connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
			return
		}

		doc := etree.NewDocument()
		//doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)

		Custromers := doc.CreateElement("Custromers")

		for _, p := range customer_map_s {
			Custromer := Custromers.CreateElement("Custromer")
			Custromer.CreateAttr("value", p.Customer_id)

			id := Custromer.CreateElement("Customer_id")
			id.CreateAttr("value", p.Customer_id)
			name := Custromer.CreateElement("Customer_name")
			name.CreateAttr("value", p.Customer_name)
			type1 := Custromer.CreateElement("Customer_type")
			type1.CreateAttr("value", p.Customer_type)
			email := Custromer.CreateElement("Customer_email")
			email.CreateAttr("value", p.Customer_email)
		}

		//doc.CreateText("/xml")

		doc.Indent(2)
		XMLString, _ := doc.WriteToString()

		fmt.Fprintf(w, XMLString)

	case "POST":

		// test_rez_slice := []CustomerStruct_xml{}
		// //var test_rez []Customer_struct
		// if err := xml.Unmarshal(xmlData, &test_rez_slice); err != nil {
		// 	panic(err)
		// }
		// fmt.Println(test_rez_slice)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		// body = []byte(`<Custromers>
		//  <Custromer value="777">
		//    <Customer_id value="777"/>
		//    <Customer_name value="Dmitry"/>
		//    <Customer_type value="Cust"/>
		//    <Customer_email value="fff@mail.ru"/>
		//  </Custromer>
		//  <Custromer value="666">
		//    <Customer_id value="666"/>
		//    <Customer_name value="Alex"/>
		//    <Customer_type value="Cust_Fiz"/>
		//    <Customer_email value="44fish@mail.ru"/>
		//  </Custromer>
		// </Custromers>`)

		doc := etree.NewDocument()
		if err := doc.ReadFromBytes(body); err != nil {
			panic(err)
		}

		var customer_map_xml = make(map[string]rootsctuct.Customer_struct)

		Custromers := doc.SelectElement("Custromers")

		for _, Custromer := range Custromers.SelectElements("Custromer") {

			Customer_struct := rootsctuct.Customer_struct{}
			//fmt.Println("CHILD element:", Custromer.Tag)
			if Customer_id := Custromer.SelectElement("Customer_id"); Customer_id != nil {
				value := Customer_id.SelectAttrValue("value", "unknown")
				Customer_struct.Customer_id = value
			}
			if Customer_name := Custromer.SelectElement("Customer_name"); Customer_name != nil {
				value := Customer_name.SelectAttrValue("value", "unknown")
				Customer_struct.Customer_name = value
			}
			if Customer_type := Custromer.SelectElement("Customer_type"); Customer_type != nil {
				value := Customer_type.SelectAttrValue("value", "unknown")
				Customer_struct.Customer_type = value
			}

			if Customer_email := Custromer.SelectElement("Customer_email"); Customer_email != nil {
				value := Customer_email.SelectAttrValue("value", "unknown")
				Customer_struct.Customer_email = value
			}

			customer_map_xml[Customer_struct.Customer_id] = Customer_struct
			// for _, attr := range Custromer.Attr {
			// 	fmt.Printf("  ATTR: %s=%s\n", attr.Key, attr.Value)
			// }
		}

		for _, p := range customer_map_xml {
			err := connector.ConnectorV.AddChangeOneRow(connector.ConnectorV.DataBaseType, p, rootsctuct.Global_settingsV)
			if err != nil {
				connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
				fmt.Println(err.Error())
			}
		}

		fmt.Fprintf(w, string(body))

	case "DELETE":

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		doc := etree.NewDocument()
		if err := doc.ReadFromBytes(body); err != nil {
			panic(err)
		}

		var customer_map_xml = make(map[string]rootsctuct.Customer_struct)

		Custromers := doc.SelectElement("Custromers")

		for _, Custromer := range Custromers.SelectElements("Custromer") {

			Customer_struct := rootsctuct.Customer_struct{}
			//fmt.Println("CHILD element:", Custromer.Tag)
			if Customer_id := Custromer.SelectElement("Customer_id"); Customer_id != nil {
				value := Customer_id.SelectAttrValue("value", "unknown")
				Customer_struct.Customer_id = value
			}
			if Customer_name := Custromer.SelectElement("Customer_name"); Customer_name != nil {
				value := Customer_name.SelectAttrValue("value", "unknown")
				Customer_struct.Customer_name = value
			}
			if Customer_type := Custromer.SelectElement("Customer_type"); Customer_type != nil {
				value := Customer_type.SelectAttrValue("value", "unknown")
				Customer_struct.Customer_type = value
			}

			if Customer_email := Custromer.SelectElement("Customer_email"); Customer_email != nil {
				value := Customer_email.SelectAttrValue("value", "unknown")
				Customer_struct.Customer_email = value
			}

			customer_map_xml[Customer_struct.Customer_id] = Customer_struct
			// for _, attr := range Custromer.Attr {
			// 	fmt.Printf("  ATTR: %s=%s\n", attr.Key, attr.Value)
			// }
		}

		for _, p := range customer_map_xml {
			err := connector.ConnectorV.DeleteOneRow(connector.ConnectorV.DataBaseType, p.Customer_id, rootsctuct.Global_settingsV)
			if err != nil {
				connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
				fmt.Println(err.Error())
			}
		}

		fmt.Fprintf(w, string(body))

	default:

		fmt.Fprintf(w, r.Method+" - This method is not implemented")

	}
}

func Test_odata_1c(w http.ResponseWriter, r *http.Request) {

	client := &http.Client{}

	req, err := http.NewRequest("GET", "http://localhost/REST_test/odata/standard.odata/Catalog_Клиенты?$format=json", nil)
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	//q := req.URL.Query()
	//q.Add("id", "id")
	//req.URL.RawQuery = q.Encode()

	fmt.Println(req.URL.String())

	resp, err := client.Do(req)

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	defer resp.Body.Close()
	resp_body, _ := ioutil.ReadAll(resp.Body)

	// fmt.Println(resp.Status)
	// fmt.Println(string(resp_body))

	var unknow_raw_json interface{}

	if err := json.Unmarshal(resp_body, &unknow_raw_json); err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	data1, _ := unknow_raw_json.(map[string]interface{})
	data1map := data1["value"]
	datadescription := data1["odata.metadata"]

	data2, _ := data1map.([]interface{})

	fmt.Println("odata.metadata : ", datadescription)
	for key, value := range data2 {
		//fmt.Println("Key:", key, "Value:", value)
		fmt.Println("----------------------Key:", key)
		valuemap := value.(map[string]interface{})
		for key2, value2 := range valuemap {
			fmt.Println("Key:", key2, "Value:", value2)
		}
	}

	fmt.Fprintf(w, string(resp_body))
}

func StratHandlers() {

	router := mux.NewRouter()

	router.HandleFunc("/", Settings)
	router.HandleFunc("/settings", Settings)

	router.HandleFunc("/rabbitMQ_1C", RabbitMQ_1C)
	router.HandleFunc("/log1C_xml", log1C_xml)
	router.HandleFunc("/api_json", Api_json)
	router.HandleFunc("/api_xml", Api_xml)

	router.HandleFunc("/test_odata_1c", Test_odata_1c)

	router.HandleFunc("/list_customer", List_customer)

	router.HandleFunc("/edit/{id:[0-9]+}", EditPage).Methods("GET")
	router.HandleFunc("/edit/{id:[0-9]+}", EditHandler).Methods("POST")
	router.HandleFunc("/delete/{id:[0-9]+}", DeleteHandler)

	router.HandleFunc("/add_change_customer", Add_change_customer)
	router.HandleFunc("/postform_add_change_customer", Postform_add_change_customer)

	http.Handle("/", router)
	fmt.Println("Server is listening...")

	http.ListenAndServe(":8181", nil)
}
