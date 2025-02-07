package main

import (    
    "log"   

    "github.com/joho/godotenv"
    "retro-vst-go/db"
)

func main() {
    // Conecta ao banco
    database, err := db.SetupDatabase()
    if err != nil {
        log.Fatalf("Falha ao configurar o banco: %v\n", err)
    }

    // 1) Roda as migrations
    if err := db.AutoMigrateDB(database); err != nil {
        log.Fatalf("Erro na migração: %v\n", err)
    }
    log.Println("Migração concluída com sucesso!")

    // 2) Pergunta se deseja inserir mock data
    // fmt.Println("Deseja inserir mock data? (s/n)")
    // var resposta string
    // fmt.Scanln(&resposta)
    // if resposta == "s" {
    //     if err := db.InsertMockData(database); err != nil {
    //         log.Fatalf("Erro ao inserir dados de teste: %v\n", err)
    //     }
    //     log.Println("Mock data inserido com sucesso!")
    // } else {
    //     fmt.Println("Dados mock NÃO inseridos.")
    // }
}
