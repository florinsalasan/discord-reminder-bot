package main

import (
    "fmt"
    "log"
    "os"

    "github.com/joho/godotenv"
)

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    APP_ID := os.Getenv("APP_ID")
    PUBLIC_KEY := os.Getenv("PUBLIC_KEY")

}
