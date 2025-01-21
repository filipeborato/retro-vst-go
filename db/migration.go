package db

import (
    "log"

    "gorm.io/gorm"

    "retro-vst-go/domain" // ajuste o import conforme seu go.mod e pastas
)

func AutoMigrateDB(db *gorm.DB) error {
    // 1) AutoMigrate: cria/atualiza as tabelas baseadas nas structs
    if err := db.AutoMigrate(
        &domain.User{},
        &domain.Product{},
        &domain.Payment{},
        &domain.Transaction{},
    ); err != nil {
        return err
    }
    log.Println("Tabelas migradas via AutoMigrate.")

    // 2) Forçar foreign_keys = ON para SQLite
    //    (No GORM, às vezes é preciso rodar esse pragma via Exec)
    if err := db.Exec("PRAGMA foreign_keys = ON;").Error; err != nil {
        return err
    }

    // 3) Criar índices (se quiser replicar o SQL "CREATE INDEX IF NOT EXISTS ...")
    // payments
    if err := db.Exec(`
        CREATE INDEX IF NOT EXISTS idx_payments_user_id 
        ON payments (user_id);
    `).Error; err != nil {
        return err
    }

    if err := db.Exec(`
        CREATE INDEX IF NOT EXISTS idx_payments_product_id 
        ON payments (product_id);
    `).Error; err != nil {
        return err
    }

    // transactions
    if err := db.Exec(`
        CREATE INDEX IF NOT EXISTS idx_transactions_user_id 
        ON transactions (user_id);
    `).Error; err != nil {
        return err
    }
    if err := db.Exec(`
        CREATE INDEX IF NOT EXISTS idx_transactions_product_id 
        ON transactions (product_id);
    `).Error; err != nil {
        return err
    }

    // 4) Criar triggers de atualização de current_balance (caso não existam)
    // Trigger: AFTER INSERT ON payments
    if err := db.Exec(`
    CREATE TRIGGER IF NOT EXISTS trg_update_balance_after_payment
    AFTER INSERT ON payments
    BEGIN
        UPDATE users
        SET current_balance = current_balance + NEW.top_up_value
        WHERE user_id = NEW.user_id;
    END;
    `).Error; err != nil {
        return err
    }

    // Trigger: AFTER INSERT ON transactions
    if err := db.Exec(`
    CREATE TRIGGER IF NOT EXISTS trg_update_balance_after_transaction
    AFTER INSERT ON transactions
    BEGIN
        UPDATE users
        SET current_balance = current_balance - NEW.transaction_value
        WHERE user_id = NEW.user_id;
    END;
    `).Error; err != nil {
        return err
    }

    log.Println("Índices e triggers criados/verificados com sucesso.")
    return nil
}
