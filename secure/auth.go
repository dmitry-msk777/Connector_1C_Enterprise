package secure

import (
	"sqlquerybuilder/config"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type MyCustomClaims struct {
	UserId uint `json:"userid"`
	jwt.StandardClaims
}

func GetMiddleWareForGin() gin.HandlerFunc {
	return func(c *gin.Context) {

		// requestPath := c.Request.URL.Path          //current request path

		tokenHeader := c.GetHeader("TokenBearer")

		if tokenHeader == "" {
			c.AbortWithStatusJSON(403, gin.H{"error": "Отсутствует токен авторизации"})
			return
		}

		//claims := jwt.MapClaims{}  // Если хотим просто выгрузить в map и перебрать открытые данные
		//claims := jwt.StandardClaims{}
		claims := &MyCustomClaims{}
		token, err := jwt.ParseWithClaims(tokenHeader, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.Conf.SecretKeyJWT), nil
		})

		if err != nil { //Malformed token, returns with http code 403 as usual
			c.AbortWithStatusJSON(403, gin.H{"error": "Неверно сформированный токен аутентификации"})
			return
		}

		if !token.Valid { //Token is invalid, maybe not signed on this server
			c.AbortWithStatusJSON(403, gin.H{"error": "Токен недействителен."})
			return
		}

		//str := fmt.Sprint(claims.UserId) //Useful for monitoring
		//fmt.Println(str)
		// ctx := context.WithValue(r.Context(), "user", tk.UserId)

		c.Next()
	}
}
