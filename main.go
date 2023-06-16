package main

import (
	"fmt"
	"log"
	"os"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("hello world")

	godotenv.Load(".env")

	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("PORT is not founded in the environment")
	}

	fmt.Println("PORT:", portString)
}