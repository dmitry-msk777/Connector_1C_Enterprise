package main

import (
	"fmt"

	connector "github.com/dmitry-msk777/Connector_1C_Enterprise/connector"
	handlers "github.com/dmitry-msk777/Connector_1C_Enterprise/handlers"

	rootsctuct "github.com/dmitry-msk777/Connector_1C_Enterprise/rootdescription"
)

func main() {

	rootsctuct.Global_settingsV.LoadSettingsFromDisk()
	connector.ConnectorV.SetSettings(rootsctuct.Global_settingsV)

	rootsctuct.LoggerCRMv.InitLog()
	connector.ConnectorV.LoggerCRM = rootsctuct.LoggerCRMv

	err := connector.ConnectorV.InitDataBase()
	if err != nil {
		connector.ConnectorV.LoggerCRM.ErrorLogger.Println(err.Error())
		fmt.Println(err.Error())
		return
	}
	//defer EngineCRMv.DatabaseSQLite.Close()

	if connector.ConnectorV.Global_settings.UseRabbitMQ {
		connector.ConnectorV.InitRabbitMQ(rootsctuct.Global_settingsV)
		//go RabbitMQ_Consumer()
	}

	handlers.StratHandlers()
}
