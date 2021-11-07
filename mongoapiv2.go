package main

import (
	"dkgosql.com/dkgosqlbooksservice/databases/mongo"
	"dkgosql.com/dkgosqlbooksservice/internal/stores"
	"github.com/gin-gonic/gin"
)

func main() {
	mongo.StartMongoAPI()
	router := gin.Default()
	router.GET("/albums", stores.CallGetBooks)
	router.GET("/albums/:id", stores.CallGetBookByID)
	router.POST("/albums", stores.CallPostBooks)

	router.Run("localhost:8100")
}
