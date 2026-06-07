package db

import (
    "fmt"
    "log"
    "os"
    "path/filepath"

    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "retro-vst-go/config"
)

func SetupDatabase() (*gorm.DB, error) {
    config.LoadEnv()    
    dbPath := config.SQLiteDBPath
    if dbPath == "" {
        dbPath = "db/retrovst.db"
    }

    // Monta path absoluto (opcional, mas ajuda a evitar problemas)
    var dbAbsolutePath string
    if filepath.IsAbs(dbPath) {
        dbAbsolutePath = dbPath
    } else {
        root, err := os.Getwd()
        if err != nil {
            return nil, fmt.Errorf("erro ao obter diretório atual: %w", err)
        }
        dbAbsolutePath = filepath.Join(root, dbPath)
    }

    // Abre conexão via GORM
    db, err := gorm.Open(sqlite.Open(dbAbsolutePath), &gorm.Config{})
    if err != nil {
        return nil, fmt.Errorf("erro ao abrir conexão com SQLite/GORM: %w", err)
    }

    // Testa a conexão
    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("erro ao obter *sql.DB: %w", err)
    }
    if err := sqlDB.Ping(); err != nil {
        return nil, fmt.Errorf("erro ao dar ping no banco: %w", err)
    }

    log.Printf("Conectado ao banco SQLite em: %s\n", dbAbsolutePath)
    return db, nil
}
