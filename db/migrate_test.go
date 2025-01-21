package db_test

import (
    "testing"
    "retro-vst-go/db"
    "retro-vst-go/domain"
    "gorm.io/gorm"
    "log"
    "os"

    "github.com/joho/godotenv"
)

func TestAutoMigrateAndSeed(t *testing.T) {
    _ = godotenv.Load() // Ignora erro

    database, err := db.SetupDatabase()
    if err != nil {
        t.Fatalf("Falha ao conectar no banco: %v", err)
    }

    // Migrate
    if err := db.AutoMigrateDB(database); err != nil {
        t.Fatalf("Erro na migração: %v", err)
    }

    // Insert mock
    if err := db.InsertMockData(database); err != nil {
        t.Fatalf("Erro ao inserir mock data: %v", err)
    }

    // Exemplo: verificar se John Doe está no banco
    var user domain.User
    result := database.Where("email = ?", "john.doe@example.com").First(&user)
    if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
        t.Fatalf("Erro ao procurar user John Doe: %v", result.Error)
    }
    log.Println("User encontrado:", user.FullName)
}
