package main

import (
	"go-graphql/data"
	"log"

	"github.com/graphql-go/handler"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {

	// simplest relay-compliant graphql server HTTP handler
	h := handler.New(&handler.Config{
		Schema: &data.Schema,
		Pretty: true,
	})
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	custom_corsmiddle := middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAcceptEncoding, "x-requested-with", "authorization"},
		AllowMethods:     []string{echo.POST, echo.OPTIONS, echo.GET},
		AllowCredentials: true,
	})

	// create graphql endpoint
	//http.Handle("/graphql", h)
	e.Any("/graphql", echo.WrapHandler(h), custom_corsmiddle)
	e.Static("/", "static")

	// serve!
	port := ":8080"
	log.Printf(`GraphQL server starting up on http://localhost%v`, port)
	e.Logger.Fatal(e.Start(port))
}
