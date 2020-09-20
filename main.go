package main

import (
	connector "github.com/dmitry-msk777/Connector_1C_Enterprise/connector"
	handlers "github.com/dmitry-msk777/Connector_1C_Enterprise/handlers"

	rootsctuct "github.com/dmitry-msk777/Connector_1C_Enterprise/rootdescription"
)

func main() {

	rootsctuct.LoggerConnV.InitLog()
	connector.ConnectorV.LoggerConn = rootsctuct.LoggerConnV

	rootsctuct.Global_settingsV.LoadSettingsFromDisk()
	err := connector.ConnectorV.SetSettings(rootsctuct.Global_settingsV)

	if err != nil {
		connector.ConnectorV.LoggerConn.ErrorLogger.Println(err.Error())
	}

	// if connector.ConnectorV.Global_settings.UseRabbitMQ {
	// 	connector.ConnectorV.InitRabbitMQ(rootsctuct.Global_settingsV)
	// 	//go RabbitMQ_Consumer()
	// }

	handlers.StratHandlers()
}
