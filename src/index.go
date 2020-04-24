package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB

// Book model
type Book struct {
	ID          int
	Name        string `gorm:"type:varchar(100);not null;"`
	Description string `gorm:"type:varchar(450);not null;"`
	AuthorID    int
	AuthorName  string `gorm:"-"`
	Editorial   string `gorm:"type:varchar(255);not null;"`
	Date        string `gorm:"type:varchar(255);not null;"`
}

// Author model
type Author struct {
	ID   int
	Name string `gorm:";type:varchar(100);not null"`
}

func main() {
	setup()
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.GET("/", homePage)
	r.GET("/books", getAll)
	r.GET("/books/:id", getByID)
	r.POST("/books", postBook)
	r.DELETE("/books/:id", deleteByID)
	r.PUT("/books/:id", updateByID)
	r.Run(":8080")
	// fmt.Println("Open localhost:8080 in your browser")
}

func getBooksWithAuthorName(books []Book) []Book {
	var author Author
	for i := 0; i < len(books); i++ {
		author, er := getAuthorByID((books)[i].ID, author)
		if er == nil {
			(books)[i].AuthorName = author.Name
		}
	}
	return books
}

// Functions for routes
func homePage(c *gin.Context) {
	var books []Book
	var authors []Author

	books, _ = getAllBooks(books)
	authors, _ = getAllAuthors(authors)
	books = getBooksWithAuthorName(books)
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"books":   books,
		"authors": authors,
	})
}

func getAll(c *gin.Context) {
	var books []Book

	var err error
	books, err = getAllBooks(books)
	if err == nil {
		if len(books) > 0 {
			books = getBooksWithAuthorName(books)
			c.JSON(200, gin.H{
				"data": books,
			})
		} else {
			c.JSON(200, gin.H{
				"message": "Sin datos.",
			})
		}
	} else {
		c.JSON(400, gin.H{
			"message": "Ah ocurrido un error",
		})
		panic(err)
	}
}

func getByID(c *gin.Context) {
	var book Book
	var author Author
	var err error
	id := c.Param("id")
	idInt, _ := strconv.Atoi(id)
	book, err = getBookByID(idInt, book)
	author, err = getAuthorByID(book.ID, author)

	if err == nil {
		c.JSON(200, gin.H{
			"data": book,
		})
	} else {
		if err.Error() == "record not found" {
			c.JSON(404, gin.H{
				"message": "Usuario no encontrado.",
			})
		} else {
			c.JSON(400, gin.H{
				"message": "Ah ocurrido un error: " + err.Error() + ".",
			})
		}
	}
}

func deleteByID(c *gin.Context) {
	var book Book
	var err error
	errMessage := ""
	id := c.Param("id")

	idInt, _ := strconv.Atoi(id)
	book, err = getBookByID(idInt, book)
	if err == nil {
		if book.ID != 0 {
			err = deleteBook(book)
			if err == nil {
				c.JSON(200, gin.H{
					"message": "Eliminado correctamente.",
				})
			} else {
				errMessage = err.Error()
			}
		} else {
			errMessage = "Usuario no encontrado."
		}
	} else {
		errMessage = err.Error()
	}
	if len(errMessage) != 0 {
		c.JSON(400, gin.H{
			"message": "Ah ocurrido un error: " + err.Error() + ".",
		})
	}
}

func updateByID(c *gin.Context) {
	var book Book
	var err error
	errMessage := ""
	id := c.Param("id")
	name := c.PostForm("name")
	description := c.PostForm("description")
	author := c.PostForm("author")
	editorial := c.PostForm("editorial")
	date := c.PostForm("date")

	idInt, _ := strconv.Atoi(id)
	book, err = getBookByID(idInt, book)
	if err == nil {
		if book.ID != 0 {
			if len(name) != 0 {
				book.Name = name
			}
			if len(description) != 0 {
				book.Description = description
			}
			if len(author) != 0 {
				var idAuthor int
				idAuthor, err = strconv.Atoi(author)
				if err == nil {
					book.AuthorID = idAuthor
				}
			}
			if len(editorial) != 0 {
				book.Editorial = editorial
			}
			if len(date) != 0 {
				book.Date = date
			}
			fmt.Println(book)
			err = updateBook(book)
			if err == nil {
				c.JSON(200, gin.H{
					"message": "Actualizado correctamente.",
				})
			} else {
				errMessage = err.Error()
			}
		} else {
			errMessage = "Usuario no encontrado."
		}
	} else {
		errMessage = err.Error()
	}
	if len(errMessage) != 0 {
		c.JSON(400, gin.H{
			"message": "Ah ocurrido un error: " + err.Error() + ".",
		})
	}
}

func postBook(c *gin.Context) {
	var err error
	name := c.PostForm("name")
	description := c.PostForm("description")
	author := c.PostForm("author")
	editorial := c.PostForm("editorial")
	date := c.PostForm("date")
	if len(name) != 0 || len(description) != 0 || len(author) != 0 || len(editorial) != 0 || len(date) != 0 {
		match, _ := regexp.MatchString("^([0]?[1-9]|[1|2][0-9]|[3][0|1])[/-]([0]?[1-9]|[1][0-2])[/-]([0-9]{4}|[0-9]{2})$", date)
		fmt.Println(match)
		if !match {
			c.JSON(400, gin.H{
				"message": "Formato de fecha inválido. Use (DD/MM/YYYY o DD-MM-YYYY)",
			})
		} else {
			var idAuthor = 0
			idAuthor, err = strconv.Atoi(author)
			book := Book{Name: name, Description: description, AuthorID: idAuthor, Editorial: editorial, Date: date}
			err = createBook(book)
			if err == nil {
				c.JSON(200, gin.H{
					"message": "Registrado correctamente",
				})
			} else {
				c.JSON(400, gin.H{
					"message": "Ocurrió un error: " + err.Error(),
				})
			}
		}

	} else {
		c.JSON(400, gin.H{
			"message": "Faltan parámetros en la solicitud",
		})
	}

}

// Functions to aperate with DB
func getAllBooks(books []Book) ([]Book, error) {
	return books, db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Find(&books).Error; err != nil {
			return err
		}
		return nil
	})
}

func getBookByID(id int, book Book) (Book, error) {
	return book, db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = " + strconv.Itoa(id)).Find(&book).Error; err != nil {
			return err
		}
		return nil
	})
}

func getAuthorByID(id int, author Author) (Author, error) {
	return author, db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = " + strconv.Itoa(id)).Find(&author).Error; err != nil {
			return err
		}
		return nil
	})
}

func getAllAuthors(authors []Author) ([]Author, error) {
	return authors, db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Find(&authors).Error; err != nil {
			return err
		}
		return nil
	})
}

func createBook(book Book) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&book).Error; err != nil {
			return err
		}
		return nil
	})
}

func deleteBook(book Book) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&book).Error; err != nil {
			return err
		}
		return nil
	})
}

func updateBook(book Book) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&book).Error; err != nil {
			return err
		}
		return nil
	})
}

func setup() {
	var err error
	server, err := sql.Open("mysql", "root:@(127.0.0.1)/")
	if err == nil {
		r, err := server.Exec("CREATE DATABASE IF NOT EXISTS books")
		fmt.Println(r)
		fmt.Println(err)
	}
	db, err = gorm.Open("mysql", "root:@(127.0.0.1)/books?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	db.DropTableIfExists(&Book{})
	db.DropTableIfExists(&Author{})

	if !db.HasTable(&Author{}) {
		db.CreateTable(&Author{})
	}
	db.AutoMigrate(&Author{})
	if !db.HasTable(&Book{}) {
		db.CreateTable(&Book{})
	}
	db.Model(&Book{}).AddForeignKey("author_id", "authors(id)", "RESTRICT", "RESTRICT")

	db.AutoMigrate(&Book{})
	seed()
}

func seed() {
	db.Save(&Author{Name: "J. K. Rowling"})
	db.Save(&Author{Name: "Carlos Muñoz"})
	db.Save(&Author{Name: "Oscar Wilde"})

	db.Save(&Book{
		Name:        "Harry Potter y la piedra filosofal",
		Description: "Primer libro de la serie literaria Harry Potter",
		AuthorID:    1,
		Editorial:   "Bloomsbury",
		Date:        "26/06/1997",
	})
	db.Save(&Book{
		Name:        "11 mentiras de las escuelas de negocios",
		Description: "Un libro que no te puedes perder si quieres tener éxito en los negocios",
		AuthorID:    2,
		Editorial:   "GRIJALBO",
		Date:        "31/12/2019",
	})
	db.Save(&Book{
		Name:        "El retrato de Dorian Gray",
		Description: "Una de las últimas obras clásicas de la novela de terror gótica con una fuerte temática faustiana",
		AuthorID:    3,
		Editorial:   "Lippincott's Monthly Magazine",
		Date:        "20/06/1890",
	})

}
