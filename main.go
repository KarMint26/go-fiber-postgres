package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/KarMint26/go-fiber-postgres/models"
	"github.com/KarMint26/go-fiber-postgres/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type Book struct {
	Author 		string		`json:"author"`
	Title 		string		`json:"title"`
	Publisher 	string		`json:"publisher"`
}

type Repository struct {
	DB *gorm.DB
}

// Create New Book Data
func (r *Repository) CreateBook(c *fiber.Ctx) error {
	book := new(Book)

	err := c.BodyParser(&book)

	if err != nil {
		c.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message":"request failed"})
			return err
	}

	createErr := r.DB.Create(&book).Error
	if createErr != nil {
		c.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message":"could not create book"})
			return err
	}

	c.Status(http.StatusOK).JSON(
		&fiber.Map{"message":"book has been added"})
	return nil
}

// Get All Books
func (r *Repository) GetBooks(c *fiber.Ctx)	error {
	bookModels := &[]models.Books{}

	err := r.DB.Find(bookModels).Error
	if err != nil {
		c.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message":"Could not get books"})
			return err
	}

	c.Status(http.StatusOK).JSON(&fiber.Map{
		"message":"books fetched successfully",
		"data": bookModels,
	})
	return nil
}

// Delete Book by Id
func (r *Repository) DeleteBook(c *fiber.Ctx) error {
	bookModel := new(models.Books)
	id := c.Params("id")
	if id == ""{
		c.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message":"id cannot be empty",
		})
		return nil
	}

	err := r.DB.Delete(bookModel, id)

	if err.Error != nil {
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message":"could not delete book",
		})
		return err.Error
	}

	c.Status(http.StatusOK).JSON(&fiber.Map{
		"message":"book delete successfully",
	})
	return nil
}

// Get Book By Id
func (r *Repository) GetBookByID(c *fiber.Ctx) error {
	id := c.Params("id")
	bookModel := &models.Books{}

	if id == "" {
		c.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id cannot be empty",
		})
		return nil
	}

	fmt.Println("the ID is", id)

	err := r.DB.Where("id = ?", id).First(bookModel).Error
	if err != nil {
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message":"could not get the book",
		})
		return err
	}

	c.Status(http.StatusOK).JSON(&fiber.Map{
		"message":"book id fetched successfully",
		"data":bookModel,
	})
	return nil
}

// Update Book
func (r *Repository) UpdateBookById(c *fiber.Ctx) error {
	bookModel := &models.Books{}
	id := c.Params("id")

	err := c.BodyParser(&bookModel)
	if err != nil {
		c.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message":"your update data not valid",
		})
		return err
	}

	if r.DB.Model(&bookModel).Where("id = ?", id).Updates(&bookModel).RowsAffected == 0 {
		c.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message":"failed to update because invalid id",
		})
		return err
	}

	c.Status(http.StatusOK).JSON(&fiber.Map{
		"message":"successfully update data",
	})
	return nil
}

func(r *Repository) SetupRoutes(app *fiber.App){
	api := app.Group("/api")
	api.Post("/create_books", r.CreateBook)
	api.Delete("/delete_book/:id", r.DeleteBook)
	api.Get("/get_books/:id", r.GetBookByID)
	api.Get("/books", r.GetBooks)
	api.Put("/update_book/:id", r.UpdateBookById)
}

func main()  {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	config := &storage.Config{
		Host: os.Getenv("DB_HOST"),
		Port: os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASS"),
		User: os.Getenv("DB_USER"),
		SSLMode: os.Getenv("DB_SSLMODE"),
		DBName: os.Getenv("DB_NAME"),
	}

	db, err := storage.NewConnection(config)

	if err != nil {
		log.Fatal("could not load the database")
	}

	err = models.MigrateBooks(db)
	if err != nil {
		log.Fatal("could not migrate db")
	}

	r := Repository{
		DB: db,
	}	
	app := fiber.New()
	r.SetupRoutes(app)
	app.Listen(":8001")
}