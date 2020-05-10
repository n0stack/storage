package main

import (
	echo "github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/n0stack/storage/chunk"
)

func main() {
	s := chunk.NewChunkStoreService("./var")

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/chunk/:chunk_id", s.ReadChunk)
	e.POST("/chunk/:chunk_id", s.WriteChunk)

	e.Logger.Fatal(e.Start(":1323"))
}
