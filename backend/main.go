package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	s3Client *s3.Client
	db *gorm.DB
	bucket string
)

func initS3() {
	
	if err := godotenv.Load("../.env"); err != nil {
		logrus.Info("No .env file found")
	}

	bucket = os.Getenv("AWS_BUCKET")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(os.Getenv("AWS_REGION")),
		config.WithBaseEndpoint(os.Getenv("AWS_ENDPOINT")),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			os.Getenv("AWS_ID"),
			os.Getenv("AWS_SECRET_KEY"),
			"",
		)),
	)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	s3Client = s3.NewFromConfig(cfg)
}

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Username     string    `gorm:"uniqueIndex;not null"`
	Email        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	CreatedAt    time.Time

	UserPhotos []UserPhoto
	Favorites  []Favorite
}

type UserPhoto struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID `gorm:"type:uuid;not null"`
	User       User      `gorm:"foreignKey:UserID"`
	URL        string    `gorm:"not null"`
	Title      string
	UploadedAt time.Time `gorm:"autoCreateTime"`
}

type Favorite struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID  uuid.UUID `gorm:"type:uuid;not null"`
	User    User      `gorm:"foreignKey:UserID"`
	PhotoID uuid.UUID `gorm:"type:uuid;not null"`
	Photo   UserPhoto `gorm:"foreignKey:PhotoID"`
	SavedAt time.Time `gorm:"autoCreateTime"`
}

type UserSchema struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserLoginSchema struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

func InitDatabase() {

	if err := godotenv.Load("../.env"); err != nil {
        log.Print("No .env file found")
    }

	host := os.Getenv("DATABASE_HOST")
	user := os.Getenv("DATABASE_USER")
	password := os.Getenv("DATABASE_PASSWORD")
	dbname := os.Getenv("DATABASE_NAME")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable", host, user, password, dbname)
	var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }

	db.AutoMigrate(&User{}, &UserPhoto{}, &Favorite{})
}

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

func getOneUser(c *fiber.Ctx) error {
	id := c.Params("id")
	
	var user User 
	if err := db.First(&user, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found",
		})
	} 

	return c.JSON(user)
}

func updateUser(c *fiber.Ctx) error {
	id := c.Params("id")

	return c.JSON(fiber.Map{
		"message": "User updated successfully",
		"id": id,
	})
	
}

func deleteUser(c *fiber.Ctx) error {
	id := c.Params("id")

	var user User

	if err := db.Delete(&user, "id = ?", id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error deleting user",
		})
	}

	return c.JSON(fiber.Map{
		"message": "User deleted successfully",
	})
}

func getUsers(c *fiber.Ctx) error {
	var users []User
	if err := db.Find(&users).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not get users"})
	}
	return c.JSON(users)
}

func getAllImages(c *fiber.Ctx) error {
	var images []UserPhoto
	if err := db.Find(&images).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not get images"})
	}

	return c.JSON(images)
}

func createImage(c *fiber.Ctx) error {

	// Get first file from form field "image":
	file, err := c.FormFile("image")
	if err != nil {
		return err
	}

	var body UserPhoto
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}



	return c.JSON(fiber.Map{
		"message": file.Filename,
	})
}

func getOneImageInfo(c *fiber.Ctx) error {
	id := c.Params("id")
	
	var image UserPhoto 
	if err := db.First(&image, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found",
		})
	} 

	return c.JSON(image)
}

// поправить роутинг для поиск фото проблема в разные id для s3 это key для бд это id
func getOneImage(c *fiber.Ctx) error {
	id := c.Params("id") // это ключ объекта в бакете

	result, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(id),
	})
	if err != nil {
		log.Fatal(err)
	}
	
	defer result.Body.Close()

	c.Set(fiber.HeaderContentType, *result.ContentType)

	return c.SendStream(result.Body)
}

func registerUser(c *fiber.Ctx) error {

	var body UserSchema
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	hash, _ := HashPassword(body.Password)

	user := User{
		ID:           uuid.New(),
		Username:     body.Username,
		Email:        body.Email,
		PasswordHash: hash,
		CreatedAt:    time.Now(),
	}

	if err := db.Create(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not create user"})
	}

	return c.JSON(user)
}

func login(c *fiber.Ctx) error {

	var body UserLoginSchema
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	var user User 
	if err := db.First(&user, "email = ?", body.Email).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found",
		})
	} 

	isPasswordValid := CheckPasswordHash(body.Password, user.PasswordHash)
	if !isPasswordValid {
		return c.Status(401).JSON(fiber.Map{
			"error": "Invalid password",
		})
	}

	// Create the Claims
	claims := jwt.MapClaims{
	 "email":  user.Email,
	 "exp":   time.Now().Add(time.Hour * 72).Unix(),
	}
	
	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
   
	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte("secret"))
	if err != nil {
	 return c.SendStatus(fiber.StatusInternalServerError)
	}
   
	return c.JSON(fiber.Map{"token": t})
}

func main() {
	InitDatabase()
	initS3()

	app := fiber.New(fiber.Config{
		AppName: "ImgHost API",
	})

	app.Get("/api/hello", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello, this is ImgHost API!",
			"version": "0.0.1",
		})
	})

	// Открытые маршруты
	app.Post("/api/auth/login", login)
	app.Post("/api/auth/register", registerUser)


	// Мидделваре
	// app.Use(jwtware.New(jwtware.Config{
	// 	SigningKey: jwtware.SigningKey{Key: []byte("secret")},
	// }))

	// Защищённые маршруты
	app.Get("/api/users", getUsers)
	app.Get("/api/users/:id", getOneUser)
	app.Patch("/api/users/:id", updateUser)
	app.Delete("/api/users/:id", deleteUser)
	// app.Get("/api/users/:id/images", getUserImages)
	// app.Get("/api/users/:id/favorites", getUserFavorites)

	app.Get("/api/images/all", getAllImages)
	app.Get("/api/images/:id/info", getOneImageInfo)
	app.Get("/api/images/:id", getOneImage)
	app.Post("/api/images/upload", createImage)
	// app.Delete("/api/images/:id", deleteImage)

	log.Println("Server is running on http://127.0.0.1:3000")
	logrus.Fatal(app.Listen("127.0.0.1:3000"))
}
