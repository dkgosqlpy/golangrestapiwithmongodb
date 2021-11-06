package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	BooksCollection   *mongo.Collection
	AuthorsCollection *mongo.Collection
	Ctx               = context.TODO()
)

type Book struct {
	Name      string `bson:"name" form:"name" binding:"required,min=3"`
	Author    string `bson:"author" form:"author" binding:"required,min=3"`
	PageCount int    `bson:"page_count" form:"count" binding:"required,min=1"`
}

type Author struct {
	FullName string `bson:"full_name"`
}

func CreateBook(b Book) (string, error) {
	result, err := BooksCollection.InsertOne(Ctx, b)
	if err != nil {
		return "0", err
	}
	return fmt.Sprintf("%v", result.InsertedID), err
}

func GetBook(id string) (Book, error) {
	var b Book
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return b, err
	}

	err = BooksCollection.
		FindOne(Ctx, bson.D{{"_id", objectId}}).
		Decode(&b)
	if err != nil {
		return b, err
	}
	return b, nil
}

func GetBooks() ([]Book, error) {
	var book Book
	var books []Book

	cursor, err := BooksCollection.Find(Ctx, bson.D{})
	if err != nil {
		defer cursor.Close(Ctx)
		return books, err
	}

	for cursor.Next(Ctx) {
		err := cursor.Decode(&book)
		if err != nil {
			return books, err
		}
		books = append(books, book)
	}

	return books, nil
}

type AuthorBooks struct {
	FullName string `bson:"full_name"`
	Books    []Book
}

func DeleteBook(id primitive.ObjectID) error {
	_, err := BooksCollection.DeleteOne(Ctx, bson.D{{"_id", id}})
	if err != nil {
		return err
	}
	return nil
}

func FindAuthorBooks(fullName string) ([]Book, error) {
	matchStage := bson.D{{"$match", bson.D{{"full_name", fullName}}}}
	lookupStage := bson.D{{"$lookup", bson.D{{"from", "books"}, {"localField", "full_name"}, {"foreignField", "author"}, {"as", "books"}}}}

	showLoadedCursor, err := AuthorsCollection.Aggregate(Ctx,
		mongo.Pipeline{matchStage, lookupStage})
	if err != nil {
		return nil, err
	}

	var a []AuthorBooks
	if err = showLoadedCursor.All(Ctx, &a); err != nil {
		return nil, err

	}
	return a[0].Books, err
}
func UpdateBook(id primitive.ObjectID, pageCount int) error {
	filter := bson.D{{"_id", id}}
	update := bson.D{{"$set", bson.D{{"page_count", pageCount}}}}
	_, err := BooksCollection.UpdateOne(
		Ctx,
		filter,
		update,
	)
	return err
}
func StartMongoAPI() {
	fmt.Println("StartMongoAPI")
	host := "127.0.0.1"
	port := "27017"
	connectionURI := "mongodb://" + host + ":" + port + "/"
	clientOptions := options.Client().ApplyURI(connectionURI)
	client, err := mongo.Connect(Ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(Ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	db := client.Database("fcmongodb")
	BooksCollection = db.Collection("books")
	AuthorsCollection = db.Collection("authors")
}

// getBooks responds with the list of all albums as JSON.
func getBooks(c *gin.Context) {
	b, err := GetBooks()
	if err != nil {
		log.Printf("Error found in getBooks : %s", err)
	}
	c.IndentedJSON(http.StatusOK, b)
}

// postBooks adds an album from JSON received in the request body.
func postBooks(c *gin.Context) {
	var b Book
	//var b = Book{Author: "Mahadevi Verma", Name: "Do bailo ki kathaa", PageCount: 30}
	if err := c.Bind(&b); err != nil {
		log.Println(b.Name)
		log.Println(b.Author)
		log.Println(b.PageCount)
		log.Printf("%#v", err)
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err})
	}
	CreateBook(b)

	if err := c.BindJSON(&b); err != nil {
		return
	}

	c.IndentedJSON(http.StatusCreated, b)
}

// getBookByID locates the album whose ID value matches the id
// parameter sent by the client, then returns that album as a response.
func getBookByID(c *gin.Context) {
	id := c.Param("id")

	a, err := GetBook(id)

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "error found"})
		return
	}

	if len(a.Author) > 0 {
		c.IndentedJSON(http.StatusOK, a)
	} else {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
	}
}

func main() {
	StartMongoAPI()
	router := gin.Default()
	router.GET("/albums", getBooks)
	router.GET("/albums/:id", getBookByID)
	router.POST("/albums", postBooks)

	router.Run("localhost:8100")
}
