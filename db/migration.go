package db

import (
    "log"

    "gorm.io/gorm"

    "retro-vst-go/domain"
)

func AutoMigrateDB(db *gorm.DB) error {
    // 1) AutoMigrate: cria/atualiza as tabelas baseadas nas structs
    if err := db.AutoMigrate(
        &domain.User{},
        &domain.Product{},
        &domain.Payment{},
        &domain.Transaction{},
        &domain.Session{},
    ); err != nil {
        return err
    }
    log.Println("Tabelas migradas via AutoMigrate.")

    // 2) Habilitar foreign_keys para SQLite
    if err := db.Exec("PRAGMA foreign_keys = ON;").Error; err != nil {
        return err
    }

    // 3) Criar índices adicionais (opcional)
    //    Ajuste para as colunas que de fato existem no seu schema.

    // payments
    if err := db.Exec(`
        CREATE INDEX IF NOT EXISTS idx_payments_user_id 
        ON payments (user_id);
    `).Error; err != nil {
        return err
    }

    // Se quiser indexar transaction_id em payments (já que substituiu product_id), use:
    if err := db.Exec(`
        CREATE INDEX IF NOT EXISTS idx_payments_transaction_id
        ON payments (transaction_id);
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

    // 4) Criar triggers para atualizar current_balance em users
    // Trigger: AFTER INSERT ON payments (adiciona top_up_value ao saldo)
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

    // Trigger: AFTER INSERT ON transactions (debita transaction_value do saldo)
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
