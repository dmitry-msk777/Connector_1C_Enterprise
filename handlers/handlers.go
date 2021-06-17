package handlers

import (
	"encoding/json"
	"net/http"
	log "sqlquerybuilder/log"
	"sqlquerybuilder/models"
	"sqlquerybuilder/store"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Ping(c *gin.Context) {

	c.String(http.StatusOK, string("Good"))

}

// GetDescription godoc
// @Summary GetDescription
// @Description Получить описание полей таблицы
// @ID get-description
// @Accept  json
// @Produce  json
// @Param TableName path string true "TableName"
// @Success 200 {object} models.ColumnsStruct
// @Header 200 {string} TokenBearer "qwerty"
// @Failure 400,404 {object} string
// @Failure 500 {object} string
// @Failure default {object} string
// @Router /getdescription/{TableName} [get]
func GetDescription(c *gin.Context) {

	TableNameParam := c.Params.ByName("TableName")

	var argsquery []interface{}
	argsquery = append(argsquery, TableNameParam)
	//queryAllColumns := "SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = $1;"

	queryAllColumns := `SELECT 
	COLUMNSTABLE.COLUMN_NAME,
	DATA_TYPE,
	IS_NULLABLE,
	CASE
					WHEN U.CONSTRAINT_NAME IS NULL THEN FALSE
					ELSE TRUE
	END AS CONSTRAINT_NAME
	FROM INFORMATION_SCHEMA.COLUMNS COLUMNSTABLE
	LEFT JOIN
				(SELECT COLUMN_NAME,
						CONSTRAINT_NAME
					FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
					WHERE TABLE_NAME = $1
					LIMIT 1) U ON COLUMNSTABLE.COLUMN_NAME = U.COLUMN_NAME
	WHERE COLUMNSTABLE.TABLE_NAME = $1`

	rows, err := store.DB.Query(queryAllColumns, argsquery...)
	if err != nil {
		c.String(http.StatusOK, "Response: %s", err.Error())
		log.Impl.Error(err.Error())
		return
	}

	ColumnsStructSlice := []models.ColumnsStruct{}
	for rows.Next() {
		var r models.ColumnsStruct
		err = rows.Scan(&r.ColumnName, &r.DataType, &r.IsNullable, &r.PrimaryKey)
		if err != nil {
			//t.Fatalf("Scan: %v", err)
			c.String(http.StatusOK, "Response: %s", err.Error())
			log.Impl.Error(err.Error())
			return
		}
		ColumnsStructSlice = append(ColumnsStructSlice, r)
	}

	rows.Close()

	byteResult, err := json.Marshal(ColumnsStructSlice)
	if err != nil {
		c.String(http.StatusOK, "Response: %s", err.Error())
		log.Impl.Error(err.Error())
		return
	}

	//fmt.Println(string(byteResult))

	c.String(http.StatusOK, string(byteResult))

}

// GetCountTable godoc
// @Summary GetCountTable
// @Description Получить количество записей в таблице
// @ID get-count-table
// @Accept  json
// @Produce  json
// @Param TableName path string true "TableName"
// @Success 200 {object} models.ColumnsStruct
// @Header 200 {string} TokenBearer "qwerty"
// @Failure 400,404 {object} string
// @Failure 500 {object} string
// @Failure default {object} string
// @Router /getcounttable/{TableName} [get]
func GetCountTable(c *gin.Context) {

	TableNameParam := c.Params.ByName("TableName")

	//var argsquery []interface{}
	//argsquery = append(argsquery, TableNameParam)

	queryString := `select COUNT(*) from ` + TableNameParam + `;`

	rows, err := store.DB.Query(queryString)
	if err != nil {
		c.String(http.StatusBadRequest, "Response: %s", err.Error())
		log.Impl.Error(err.Error())
		return
	}

	var count int

	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			//t.Fatalf("Scan: %v", err)
			c.String(http.StatusBadRequest, "Response: %s", err.Error())
			log.Impl.Error(err.Error())
			return
		}
	}

	rows.Close()

	c.String(http.StatusOK, strconv.Itoa(count))
}

func ClearTable(c *gin.Context) {

	TableNameParam := c.Params.ByName("TableName")

	//QueryCompose := "CREATE TABLE IF NOT EXISTS " + TableDescription.TableName + " (id BIGINT NOT NULL PRIMARY KEY)"
	QueryString := "DELETE FROM " + TableNameParam + ";"

	//fmt.Println(QueryCompose)

	_, err := store.DB.Exec(QueryString)
	if err != nil {
		c.String(http.StatusBadRequest, "Response: %s", err.Error())
		log.Impl.Error(err.Error())
		return
	}

	c.String(http.StatusOK, "Status OK, query for Pg %s", QueryString)

}

func StressTesting(c *gin.Context) {

	//var a [10000000][]int

	var count int64
	for i := 0; i < 10000000; i++ {

		//var stack [1024]byte
		//heap := make([]byte, 64*1024)
		//heap := make([]byte, 1024)
		//_, _ = stack, heap

		//s := make([]int32, 1000)
		//_ = s

		//a[i] = make([]int, 1024)
		// Это более мение норм вариант
		//a[i] = make([]int, 1)

		count = +1
	}

	_ = count

	c.String(http.StatusOK, "Good")
}
