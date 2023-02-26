package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Book struct {
	Id     string `json:"id" binding:"required"`
	Title  string `json:"title" binding:"required"`
	Author string `json:"author" binding:"required"`
}

var Err error
var Client *mongo.Client
var Collection *mongo.Collection

func GetHelloWorld(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Hello World"})
}

//Retrieve all books
func GetAllBooks(c *gin.Context) {

	filter := bson.M{}
	cursor, err := Collection.Find(context.TODO(), filter)
	checkErr(err)
	var results []Book
	err = cursor.All(context.TODO(), &results)
	checkErr(err)
	// fmt.Println(results)
	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Retrieve all books", "books": results})

}

//Retrieve each book by id
func GetBookId(c *gin.Context) {
	id := c.Param("id")

	filter := bson.M{"id": id}
	cursor, err := Collection.Find(context.TODO(), filter)
	checkErr(err)
	var results []Book
	err = cursor.All(context.TODO(), &results)
	if results != nil {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Retrieve book", "books": results})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Do not have book in database"})
	}

}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

//add
func AddBook(c *gin.Context) {
	var b Book
	c.BindJSON(&b)
	filter := bson.M{"id": b.Id}
	cursor, err := Collection.Find(context.TODO(), filter)
	checkErr(err)
	var seachBook []Book
	err = cursor.All(context.TODO(), &seachBook)
	if seachBook == nil {
		//Add book
		newBook := Book{Id: b.Id, Title: b.Title, Author: b.Author}
		_, err = Collection.InsertOne(context.TODO(), newBook)
		checkErr(err)

		// check book exist
		cursor, err = Collection.Find(context.TODO(), filter)
		checkErr(err)
		var results []Book
		err = cursor.All(context.TODO(), &results)
		if results != nil {
			c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Added book", "books": results})
		} else {
			c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Cannnot add book"})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Book's id is duplicate"})
	}

}

//update
func UpdateBookData(c *gin.Context) {
	var b Book
	c.BindJSON(&b)
	filter := bson.M{"id": b.Id}
	cursor, err := Collection.Find(context.TODO(), filter)
	checkErr(err)
	var seachBook []Book
	err = cursor.All(context.TODO(), &seachBook)
	if seachBook == nil {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Book does not exist"})
		return
	}

	update := bson.D{{"$set", bson.D{
		{"title", b.Title},
		{"author", b.Author},
	}}}
	_, err = Collection.UpdateOne(context.TODO(), filter, update)
	// check book update

	cursor, err = Collection.Find(context.TODO(), filter)
	checkErr(err)
	var results []Book
	err = cursor.All(context.TODO(), &results)
	if results != nil {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Updated book's data", "book": results})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Cannnot update book's data"})
	}

}

//delete
func DeleteBookId(c *gin.Context) {
	var b Book
	c.BindJSON(&b)

	// เรียกดู database
	filter := bson.M{"id": b.Id}

	cursor, err := Collection.Find(context.TODO(), filter)
	checkErr(err)
	var seachBook []Book
	err = cursor.All(context.TODO(), &seachBook)
	if seachBook == nil {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Book does not exist"})
		return
	}
	_, err = Collection.DeleteOne(context.TODO(), filter)
	checkErr(err)

	//check book exist
	cursor, err = Collection.Find(context.TODO(), filter)
	checkErr(err)
	var seachCheck []Book
	err = cursor.All(context.TODO(), &seachCheck)
	fmt.Println(seachCheck)
	if seachCheck == nil {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Book has already deleted"})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Book exist", "book": seachCheck})
	}
}

//login
func loginHandler(c *gin.Context) {
	// implement login logic here

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(5 * time.Minute).Unix(),
	})
	ss, err := token.SignedString([]byte("MatusLhongpol"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"token": ss,
	})
}

//validate token
func validateToken(token string) error {
	//jwt.Parse use for read token
	validate, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte("MatusLhongpol"), nil //return signature
	})
	fmt.Println(validate)
	return err
}

//middleware
func authorizationMiddleware(c *gin.Context) {
	s := c.Request.Header.Get("Authorization")

	token := strings.TrimPrefix(s, "Bearer ")

	if err := validateToken(token); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "Please login"})
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
}

func main() {

	Client, Err = mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	checkErr(Err)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	Err = Client.Connect(ctx)
	checkErr(Err)
	Collection = Client.Database("mydb").Collection("bookstore")

	r := gin.Default()

	r.POST("/login", loginHandler)
	protected := r.Group("/", authorizationMiddleware)

	protected.GET("/", GetHelloWorld)
	protected.GET("/allbooks", GetAllBooks)
	protected.GET("/book/:id", GetBookId)
	protected.POST("/book/add", AddBook)
	protected.PUT("/book/update", UpdateBookData)
	protected.DELETE("/book/delete", DeleteBookId)

	r.Run(":3000") // listen and serve on 0.0.0.0:3000 (for windows "localhost:8080")
}
