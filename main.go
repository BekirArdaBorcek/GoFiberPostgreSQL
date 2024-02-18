package main

import (
	"context"
	"log"
	"strconv"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	App := fiber.New(fiber.Config{
		Prefork:      false,
		AppName:      "App",
		ServerHeader: "App",
		JSONEncoder:  sonic.Marshal,
		JSONDecoder:  sonic.Unmarshal,
	})

	Pgx, DatabaseError := pgxpool.New(context.Background(), "postgres://postgres:6563001109aA@localhost:5432/Test?sslmode=disable")
	if DatabaseError != nil {
		panic(DatabaseError)
	}
	defer Pgx.Close()

	App.Get("/", func(Context *fiber.Ctx) error {
		return GetAllUser(Context, Pgx)
	})
	App.Get("/:id", func(Context *fiber.Ctx) error {
		return GetUserByID(Context, Pgx)
	})
	App.Post("/", func(Context *fiber.Ctx) error {
		return CreateUser(Context, Pgx)
	})
	App.Put("/:id", func(Context *fiber.Ctx) error {
		return UpdateUser(Context, Pgx)
	})
	App.Delete("/:id", func(Context *fiber.Ctx) error {
		return DeleteUser(Context, Pgx)
	})

	log.Fatal(App.Listen(":6563"))
}

func GetAllUser(Context *fiber.Ctx, Pgx *pgxpool.Pool) error {
	Query, QueryError := Pgx.Query(context.Background(), "SELECT * FROM users")
	if QueryError != nil {
		return QueryError
	}
	defer Query.Close()

	var Users []User

	for Query.Next() {
		var ForUser User

		if ScanError := Query.Scan(&ForUser.ID, &ForUser.Name, &ForUser.Email); ScanError != nil {
			return ScanError
		}
		Users = append(Users, ForUser)
	}

	Context.Status(fiber.StatusOK)
	return Context.JSON(Users)
}

func GetUserByID(Context *fiber.Ctx, Pgx *pgxpool.Pool) error {

	GetUserByID := Context.Params("id")
	ID, Error := strconv.Atoi(GetUserByID)
	if Error != nil {
		Context.Status(fiber.StatusBadRequest)
		return Context.JSON(fiber.Map{"message": "Geçersiz ID"})
	}

	var User User
	Error = Pgx.QueryRow(context.Background(), "SELECT * FROM users WHERE id = $1", ID).Scan(&User.ID, &User.Name, &User.Email)
	if Error != nil {

		if Error.Error() == "no rows in result set" {
			Context.Status(fiber.StatusNotFound)
			return Context.JSON(fiber.Map{"message": "Kullanıcı Bulunamadı"})
		}
		return Error
	}
	Context.Status(fiber.StatusOK)
	return Context.JSON(User)
}

func CreateUser(Context *fiber.Ctx, Pgx *pgxpool.Pool) error {
	var User User
	if Error := Context.BodyParser(&User); Error != nil {
		Context.Status(fiber.StatusBadRequest)
		return Context.JSON(fiber.Map{"message": "Geçersiz veri"})
	}
	_, Error := Pgx.Exec(context.Background(), "INSERT INTO users (name, email) VALUES ($1, $2)", User.Name, User.Email)
	if Error != nil {
		return Error
	}
	Context.Status(fiber.StatusCreated)
	return Context.JSON(User)
}

func UpdateUser(Context *fiber.Ctx, Pgx *pgxpool.Pool) error {
	GetUserByID := Context.Params("id")
	ID, Error := strconv.Atoi(GetUserByID)
	if Error != nil {
		Context.Status(fiber.StatusBadRequest)
		return Context.JSON(fiber.Map{"message": "Geçersiz ID"})
	}
	var User User
	if Error := Context.BodyParser(&User); Error != nil {
		Context.Status(fiber.StatusBadRequest)
		return Context.JSON(fiber.Map{"message": "Geçersiz veri"})
	}
	_, Error = Pgx.Exec(context.Background(), "UPDATE users SET name = $1, email = $2 WHERE id = $3", User.Name, User.Email, ID)
	if Error != nil {
		return Error
	}
	Context.Status(fiber.StatusOK)
	return Context.JSON(User)
}

func DeleteUser(Context *fiber.Ctx, Pgx *pgxpool.Pool) error {
	GetUserByID := Context.Params("id")
	ID, Error := strconv.Atoi(GetUserByID)
	if Error != nil {
		Context.Status(fiber.StatusBadRequest)
		return Context.JSON(fiber.Map{"message": "Geçersiz ID"})
	}
	_, Error = Pgx.Exec(context.Background(), "DELETE FROM users WHERE id = $1", ID)
	if Error != nil {
		Context.Status(fiber.StatusInternalServerError)
		return Context.JSON(fiber.Map{"message": "Kullanıcı silinirken bir hata oluştu"})
	}
	Context.Status(fiber.StatusOK)
	return Context.JSON(fiber.Map{"message": "Kullanıcı başarıyla silindi"})
}
