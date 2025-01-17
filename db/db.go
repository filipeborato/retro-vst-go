// db/db.go

package db

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "path/filepath"

    _ "github.com/mattn/go-sqlite3"
)

// SetupDatabase lê a variável relativa em SQLITE_DB_PATH,
// e monta o caminho absoluto usando o diretório atual do projeto.
func SetupDatabase() (*sql.DB, error) {
    // Lê o valor da variável de ambiente
    dbRelativePath := os.Getenv("SQLITE_DB_PATH")
    if dbRelativePath == "" {
        // Se não estiver definido no .env, usamos um padrão
        dbRelativePath = "db/retrovst.db"
    }

    // Pega o diretório de trabalho atual
    projectRoot, err := os.Getwd()
    if err != nil {
        return nil, fmt.Errorf("erro ao obter diretorio do projeto: %w", err)
    }

    // Monta o caminho absoluto
    dbPath := filepath.Join(projectRoot, dbRelativePath)

    // Abre a conexão com SQLite
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, fmt.Errorf("erro ao abrir conexao com o banco: %w", err)
    }

    // Testa a conexão
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("erro ao dar ping no banco: %w", err)
    }

    log.Printf("Conectado ao banco SQLite em: %s\n", dbPath)
    return db, nil
}
