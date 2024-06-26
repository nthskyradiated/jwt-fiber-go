package main

import (
	"time"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/jwt/v3"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"github.com/nthskyradiated/jwt-fiber-go/db"
)

type SignupRequest struct {
	Name string
	Email string
	Password string
}
type LoginRequest struct {
	Email string
	Password string
}
func main(){
	app := fiber.New()

	engine, err := db.CreateDBEngine()
	if err !=nil {
		panic(err)
	}
	app.Post("/signup", func(c *fiber.Ctx) error {
		req := new(SignupRequest)
		if err := c.BodyParser(req); err != nil {
			return err
		}
		if req.Name == "" || req.Email == "" || req.Password =="" {
			return fiber.NewError(fiber.StatusBadRequest, "invalid credentials")
		}
			//save info
			hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if  err != nil {
				return err
			}
			user := &db.User {
				Name: req.Name,
				Email: req.Email,
				Password: string(hash),
			}

			_, err = engine.Insert(user)
			if err != nil {
				return err
			}
			token, exp, err := createToken(*user)
			if err != nil {
				return err
			}
			return c.JSON(fiber.Map{"token": token, "exp": exp, "user": user})
	})
	app.Post("/login", func(c *fiber.Ctx) error {
		req := new(LoginRequest)
		if err := c.BodyParser(req); err != nil {
			return err
		}
		if req.Email == "" || req.Password =="" {
			return fiber.NewError(fiber.StatusBadRequest, "invalid credentials")
		}
		user := new(db.User)
		has, err := engine.Where("email= ?", req.Email).Desc("id").Get(user)
		if err != nil {
			return err
		}
		if !has {
			return fiber.NewError(fiber.StatusBadRequest, "invalid login")
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			return err
		}

		token, exp, err := createToken(*user)
		if err != nil {
			return err
		}
		return c.JSON(fiber.Map{"token": token, "exp": exp, "user": user})

	})

	private := app.Group("/private")
	private.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte("secret"),
	}))
	private.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"path": "private",
		})
	})
	public := app.Group("/public")
	public.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"path": "public",
		})
	})
	if err := app.Listen(":3000"); err != nil {
		panic(err)
	}
}

func createToken(user db.User) (string, int64, error) {
	exp := time.Now().Add(time.Minute * 30).Unix()
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.Id
	claims["exp"] = exp
	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return "",0,err
	}
	return t, exp, nil
}