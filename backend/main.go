package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

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

func InitDatabase() {
	var err error
	DB, err = gorm.Open(sqlite.Open("app.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = DB.AutoMigrate(&User{}, &UserPhoto{}, &Favorite{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
}

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

func createUser(c *fiber.Ctx) error {
	type Request struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var body Request
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

	if err := DB.Create(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not create user"})
	}

	return c.JSON(user)
}

func getUsers(c *fiber.Ctx) error {
	var users []User
	if err := DB.Find(&users).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not get users"})
	}
	return c.JSON(users)
}

func main() {
	InitDatabase()

	app := fiber.New()

	app.Get("/api/hello", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello from Go backend!",
		})
	})

	app.Post("/api/users", createUser)
	app.Get("/api/users", getUsers)
	app.Get("/api/users/:id", getOneUser)
	app.Patch("/api/users/:id", updateUser)
	app.Delete("/api/users/:id", deleteUser)

	log.Println("Server is running on http://127.0.0.1:3000")
	logrus.Fatal(app.Listen("127.0.0.1:3000"))
}
