package handlers

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gorilla/mux"

	connector "github.com/dmitry-msk777/Connector_1C_Enterprise/connector"
	rootsctuct "github.com/dmitry-msk777/Connector_1C_Enterprise/rootdescription"

	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"

	"github.com/beevik/etree"

	_ "github.com/dmitry-msk777/Connector_1C_Enterprise/docs"
	httpSwagger "github.com/swaggo/http-swagger" // http-swagger middleware
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

		if r.FormValue("UseTelegram") == "on" {
			rootsctuct.Global_settingsV.UseTelegram = true
		} else {
			rootsctuct.Global_settingsV.UseTelegram = false
		}

		rootsctuct.Global_settingsV.TelegramAPIKey = r.FormValue("TelegramAPIKey")

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

	// Тестировалаось на 6.8.6

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

		start := time.Now()

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		fmt.Println("Size byte : ", binary.Size(body))

		// Возникают проблемы при загрузке файла размеров в 1 GB это 100 000 записей журнала
		// Log1C_slice, err := connector.ConnectorV.ParseXMLFrom1C(body)
		// if err != nil {
		// 	connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
		// 	fmt.Fprintf(w, err.Error())
		// }

		// fmt.Printf("len=%d cap=%d %v\n", len(Log1C_slice), cap(Log1C_slice))

		// Можно разпознать XML по сайту и получить похожую структуру слайзов в EventLog1C.Event
		// Сайт генерации структуры по файлу https://www.onlinetool.io/xmltogo/
		var EventLog1C rootsctuct.EventLog1C

		err = xml.Unmarshal(body, &EventLog1C)
		if err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}

		// for _, Event := range EventLog1C.Event {
		// 	fmt.Println(Event)
		// }

		fmt.Printf("len=%d cap=%d", len(EventLog1C.Event), cap(EventLog1C.Event))

		duration := time.Since(start)
		fmt.Println(duration)

		// Это варинат загрузки по одной записи из библиотеке представляющей XML как DOM
		// err = connector.ConnectorV.SendInElastichSearchOld(Log1C_slice)

		// Этот вариант запись по одной записи без Bulk
		//err = connector.ConnectorV.SendInElastichSearchNew(EventLog1C.Event)

		// Тут нет многопоточности при записи, но она есть в хэндлере с сжатием
		err = connector.ConnectorV.SendInElastichBulk(EventLog1C.Event)

		if err != nil {
			connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
			return
		}

		duration2 := time.Since(start)
		fmt.Println(duration2)

		fmt.Fprintf(w, "Succeed!")

	}
}

// API JSON GET godoc
// @Summary Exchange Customer
// @Description Get-Set Customer
// @ID Exchange-Customer
// @Accept  json
// @Produce  json
// @Param id_customer query string false "id_customer"
// @Success 200 {array} rootsctuct.Customer_struct
// @Header 200 {string} Token "qwerty"
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Failure 500 {object} string
// @Router /api_json [post]
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

// API JSON godoc
// @Summary Get all Customer
// @Description Get all Customer
// @ID Get-all-Customer
// @Accept  json
// @Produce  json
// @Success 200 {array} rootsctuct.Customer_struct
// @Header 200 {string} Token "qwerty"
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Failure 500 {object} string
// @Router /list_customer [get]
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

	// Тут парсим неопределенный JSON
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

	// Можно переложить результат в какую-нибудь структуру пример ниже

	// example of type definition
	// switch vv := v.(type) {
	// case string:
	//     fmt.Printf("%s => (string) %q\n", kn, vv)
	// case bool:
	//     fmt.Printf("%s => (bool) %v\n", kn, vv)
	// case float64:
	//     fmt.Printf("%s => (float64) %f\n", kn, vv)
	// case map[string]interface{}:
	//     fmt.Printf("%s => (map[string]interface{}) ...\n", kn)
	//     iterMap(vv, kn)
	// case []interface{}:
	//     fmt.Printf("%s => ([]interface{}) ...\n", kn)
	//     iterSlice(vv, kn)
	// default:
	//     fmt.Printf("%s => (unknown?) ...\n", kn)
	// }

	// Тут идет преобразования с определенной структурой для документа. Выше для неопределенного JSON
	var Odata1C rootsctuct.Odata1C

	if err := json.Unmarshal(resp_body, &Odata1C); err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	fmt.Fprintf(w, string(resp_body))

}

func Test(w http.ResponseWriter, r *http.Request) {

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

	case "POST", "PUT":

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

		// for _, p := range customer_map_json {
		// 	err := connector.ConnectorV.AddChangeOneRow(connector.ConnectorV.DataBaseType, p, rootsctuct.Global_settingsV)
		// 	if err != nil {
		// 		connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
		// 		fmt.Println(err.Error())
		// 	}
		// }

		fmt.Fprintf(w, string(body))

	default:

		fmt.Fprintf(w, r.Method+" - This method is not implemented")

	}

	fmt.Fprintf(w, "test")
}

func readZipFile(zf *zip.File) ([]byte, error) {
	f, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

func ParseXMLThroughDecored(ByteData []byte) ([]rootsctuct.Event1C, error) {

	input := bytes.NewReader(ByteData)
	decoder := xml.NewDecoder(input)

	Events := make([]rootsctuct.Event1C, 0)
	var Event1C rootsctuct.Event1C

	for {

		tok, tokenErr := decoder.Token()

		if tokenErr != nil && tokenErr != io.EOF {
			break
		} else if tokenErr == io.EOF {
			break
		}

		if tok == nil {
			//break
		}

		switch tok := tok.(type) {
		case xml.StartElement:
			if tok.Name.Local == "Event" {
				if err := decoder.DecodeElement(&Event1C, &tok); err != nil {

				}
				Events = append(Events, Event1C)
			}

		case xml.CharData:
			//data := strings.TrimSpace(string(tok))
			fmt.Println(tok)
		case xml.EndElement:
			fmt.Println(tok.Name.Local)
		}

	}

	return Events, nil

}

func ConverXMLLog1C(Event1CExtended []rootsctuct.Event1CExtended) ([]rootsctuct.Event1C, error) {

	NumCPU := runtime.NumCPU()
	// runtime.GOMAXPROCS(NumCPU)

	var divided [][]rootsctuct.Event1CExtended

	chunkSize := (len(Event1CExtended) + NumCPU - 1) / NumCPU

	for i := 0; i < len(Event1CExtended); i += chunkSize {
		end := i + chunkSize

		if end > len(Event1CExtended) {
			end = len(Event1CExtended)
		}

		divided = append(divided, Event1CExtended[i:end])
	}

	var Events []rootsctuct.Event1C
	mu := &sync.Mutex{}

	var wg sync.WaitGroup
	for _, sliceRow := range divided {
		wg.Add(1)
		go func(sliceRow []rootsctuct.Event1CExtended, mu *sync.Mutex) error {
			defer wg.Done()

			for _, Slice1 := range sliceRow {

				var Event rootsctuct.Event1C

				Event.Text = Slice1.Text
				Event.Level = Slice1.Level
				Event.Date = Slice1.Date
				Event.ApplicationName = Slice1.ApplicationName
				Event.ApplicationPresentation = Slice1.ApplicationPresentation
				Event.Event = Slice1.Event
				Event.EventPresentation = Slice1.EventPresentation
				Event.User = Slice1.User
				Event.UserName = Slice1.UserName
				Event.Computer = Slice1.Computer
				Event.Metadata = Slice1.Metadata
				Event.MetadataPresentation = Slice1.MetadataPresentation
				Event.Comment = Slice1.Comment

				Event.DataPresentation = Slice1.DataPresentation
				Event.TransactionStatus = Slice1.TransactionStatus
				Event.TransactionID = Slice1.TransactionID
				Event.Connection = Slice1.Connection
				Event.Session = Slice1.Session
				Event.ServerName = Slice1.ServerName
				Event.Port = Slice1.Port
				Event.SyncPort = Slice1.SyncPort

				var Event1CExtendedData rootsctuct.Event1CExtendedData

				bytejsone, err := json.Marshal(&Event1CExtendedData)
				if err != nil {
					return err
				}

				Event.Data = string(bytejsone)

				mu.Lock()
				Events = append(Events, Event)
				mu.Unlock()

			}

			return nil
		}(sliceRow, mu)
	}
	wg.Wait()

	return Events, nil
}

func log1C_zip(w http.ResponseWriter, r *http.Request) {

	// Тестировалаось на 6.8.6

	if r.Method == "GET" {

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
		//fmt.Fprintf(w, string(JsonString))

		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		json.NewEncoder(gz).Encode(JsonString)
		gz.Close()

	} else {

		start := time.Now()

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		fmt.Println("Size byte : ", binary.Size(body))

		zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		if err != nil {
			log.Fatal(err)
		}

		var unzippedFileBytes []byte
		// Read all the files from zip archive
		for _, zipFile := range zipReader.File {
			fmt.Println("Reading file:", zipFile.Name)
			unzippedFileBytes, err = readZipFile(zipFile)
			if err != nil {
				log.Println(err)
				continue
			}

			_ = unzippedFileBytes // this is unzipped file bytes
		}

		fmt.Println("Size byte unzip : ", binary.Size(unzippedFileBytes))

		// Этот блок парсинг через декодер, который дольше получается чем стандартный xml.Unmarshal(
		// Event, err := ParseXMLThroughDecored(unzippedFileBytes)
		// if err != nil {
		// 	connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
		// 	fmt.Fprintf(w, err.Error())
		// 	return
		// }
		// fmt.Printf("len=%d cap=%d %v\n", len(Event), cap(Event))

		var EventLog1C rootsctuct.EventLog1C

		err = xml.Unmarshal(unzippedFileBytes, &EventLog1C)
		if err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}

		fmt.Printf("len=%d cap=%d", len(EventLog1C.Event), cap(EventLog1C.Event))

		// Тут получаем вложенное поле Data в виде структуры в формате JSON
		// Использовать var EventLog1C rootsctuct.EventLog1CExtended
		//_, err = ConverXMLLog1C(EventLog1C.Event)

		if err != nil {
			connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
			return
		}

		duration := time.Since(start)
		fmt.Println(duration)

		err = connector.ConnectorV.SendInElastichBulkGOroutines(EventLog1C.Event)

		if err != nil {
			connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
			return
		}

		duration2 := time.Since(start)
		fmt.Println(duration2)

		fmt.Fprintf(w, "Succeed!")

	}
}

// @title Swagger API Connector for 1C Enterprise
// @version 1.0

// @description API description
// @termsOfService http://swagger.io/terms/

// @contact.name Dmitry
// @contact.url https://github.com/dmitry-msk777/Connector_1C_Enterprise

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8181
// @BasePath /v2

func StratHandlers() {

	router := mux.NewRouter()

	router.HandleFunc("/", Settings)
	router.HandleFunc("/settings", Settings)

	// https://github.com/swaggo/swag
	// swag init -g handlers/handlers.go
	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8181/swagger/doc.json"), //The url pointing to API definition
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("#swagger-ui"),
	))

	router.HandleFunc("/rabbitMQ_1C", RabbitMQ_1C)
	router.HandleFunc("/log1C_xml", log1C_xml)
	router.HandleFunc("/log1C_zip", log1C_zip)

	router.HandleFunc("/api_json", Api_json)
	router.HandleFunc("/api_xml", Api_xml)

	router.HandleFunc("/test", Test)

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
