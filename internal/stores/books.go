package stores

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	conmongo "dkgosql.com/dkgosqlbooksservice/databases/mongo"
	"dkgosql.com/dkgosqlbooksservice/internal/models"
)

func CreateBook(b models.Book) (string, error) {
	result, err := conmongo.BooksCollection.InsertOne(conmongo.Ctx, b)
	if err != nil {
		return "0", err
	}
	return fmt.Sprintf("%v", result.InsertedID), err
}

func GetBook(id string) (models.Book, error) {
	var b models.Book
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return b, err
	}

	err = conmongo.BooksCollection.
		FindOne(conmongo.Ctx, bson.D{{"_id", objectId}}).
		Decode(&b)
	if err != nil {
		return b, err
	}
	return b, nil
}

func GetBooks() ([]models.Book, error) {
	var book models.Book
	var books []models.Book

	cursor, err := conmongo.BooksCollection.Find(conmongo.Ctx, bson.D{})
	if err != nil {
		defer cursor.Close(conmongo.Ctx)
		return books, err
	}

	for cursor.Next(conmongo.Ctx) {
		err := cursor.Decode(&book)
		if err != nil {
			return books, err
		}
		books = append(books, book)
	}

	return books, nil
}

func DeleteBook(id primitive.ObjectID) error {
	_, err := conmongo.BooksCollection.DeleteOne(conmongo.Ctx, bson.D{{"_id", id}})
	if err != nil {
		return err
	}
	return nil
}

func FindAuthorBooks(fullName string) ([]models.Book, error) {
	matchStage := bson.D{{"$match", bson.D{{"full_name", fullName}}}}
	lookupStage := bson.D{{"$lookup", bson.D{{"from", "books"}, {"localField", "full_name"}, {"foreignField", "author"}, {"as", "books"}}}}

	showLoadedCursor, err := conmongo.AuthorsCollection.Aggregate(conmongo.Ctx,
		mongo.Pipeline{matchStage, lookupStage})
	if err != nil {
		return nil, err
	}

	var a []models.AuthorBooks
	if err = showLoadedCursor.All(conmongo.Ctx, &a); err != nil {
		return nil, err

	}
	return a[0].Books, err
}
func UpdateBook(id primitive.ObjectID, pageCount int) error {
	filter := bson.D{{"_id", id}}
	update := bson.D{{"$set", bson.D{{"page_count", pageCount}}}}
	_, err := conmongo.BooksCollection.UpdateOne(
		conmongo.Ctx,
		filter,
		update,
	)
	return err
}

// getBooks responds with the list of all albums as JSON.
func CallGetBooks(c *gin.Context) {
	b, err := GetBooks()
	if err != nil {
		log.Printf("Error found in getBooks : %s", err)
	}
	c.IndentedJSON(http.StatusOK, b)
}

// postBooks adds an album from JSON received in the request body.
func CallPostBooks(c *gin.Context) {
	var b models.Book
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
func CallGetBookByID(c *gin.Context) {
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
