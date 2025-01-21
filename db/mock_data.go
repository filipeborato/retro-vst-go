package db

import (
    "log"

    "gorm.io/gorm"

    "retro-vst-go/domain" // Ajuste o import conforme seu go.mod
)

// InsertMockData insere dados de teste (usuários, produtos, payments, transactions).
// É opcional e pensado para ambientes de desenvolvimento / testes.
func InsertMockData(db *gorm.DB) error {
    log.Println("Inserindo dados de teste...")

    // 1) Users
    users := []domain.User{
        {FullName: "John Doe", Email: "john.doe@example.com", PasswordHash: "hashsenha123"},
        {FullName: "Jane Smith", Email: "jane.smith@example.com", PasswordHash: "hashsenha456"},
    }
    if err := db.Create(&users).Error; err != nil {
        return err
    }

    // 2) Products
    products := []domain.Product{
        {ProductName: "TheFunction", Description: "Plugin de espacialidade avançada", Price: 0.20},
        {ProductName: "TheWave", Description: "Plugin da ondinha", Price: 0.20},
    }
    if err := db.Create(&products).Error; err != nil {
        return err
    }

    // (Opcional) Se quiser capturar os IDs gerados
    // por ex: users[0].UserID, products[0].ProductID, etc.

    // 3) Payments (exemplo de recarga)
    // Supondo que user 1 recarregue R$50, user 2 recarregue R$100
    payments := []domain.Payment{
        {
            UserID:            users[0].UserID, // John
            ProductID:         nil,
            ExternalPaymentID: "EXT-ABC-123",
            SupplierName:      "PayPal",
            TopUpValue:        50.00,
            BalanceAfterTopUp: 50.00,
        },
        {
            UserID:            users[1].UserID, // Jane
            ProductID:         nil,
            ExternalPaymentID: "EXT-XYZ-789",
            SupplierName:      "Stripe",
            TopUpValue:        100.00,
            BalanceAfterTopUp: 100.00,
        },
    }
    if err := db.Create(&payments).Error; err != nil {
        return err
    }

    // 4) Transactions (débito) - John compra um dos produtos
    transactions := []domain.Transaction{
        {
            UserID:          users[0].UserID, // John
            ProductID:       products[0].ProductID, // "TheFunction"
            TransactionValue: 10.00,
        },
        {
            UserID:          users[1].UserID, // Jane
            ProductID:       products[1].ProductID, // "TheWave"
            TransactionValue: 20.00,
        },
        {
            UserID:          users[0].UserID, // John again
            ProductID:       products[1].ProductID, // "TheWave"
            TransactionValue: 20.00,
        },
    }
    if err := db.Create(&transactions).Error; err != nil {
        return err
    }

    log.Println("Dados de teste inseridos com sucesso.")
    return nil
}
