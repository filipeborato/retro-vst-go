-- ============================================================
-- Habilitar verificação de Foreign Keys no SQLite
-- ============================================================
PRAGMA foreign_keys = ON;

-- ============================================================
-- Table: users
-- ============================================================
CREATE TABLE IF NOT EXISTS users (
    user_id INTEGER PRIMARY KEY AUTOINCREMENT,
    full_name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    current_balance DECIMAL(10, 2) NOT NULL DEFAULT 0,  -- Saldo atual do usuário
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================
-- Table: products
-- ============================================================
CREATE TABLE IF NOT EXISTS products (
    product_id INTEGER PRIMARY KEY AUTOINCREMENT,
    product_name TEXT NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================
-- Table: payments
-- Guarda as recargas de saldo realizadas pelo usuário.
-- ============================================================
CREATE TABLE IF NOT EXISTS payments (
    payment_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    product_id INTEGER,  -- Se a recarga estiver vinculada a um produto
    external_payment_id TEXT NOT NULL,  -- ID de pagamento fornecido pelo gateway/fornecedor
    supplier_name TEXT NOT NULL,        -- Nome do fornecedor (ex: PayPal, Stripe, etc.)
    top_up_value DECIMAL(10, 2) NOT NULL DEFAULT 0,        -- Valor da recarga
    balance_after_top_up DECIMAL(10, 2) NOT NULL DEFAULT 0, -- Saldo do usuário após essa recarga
    payment_date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (product_id) REFERENCES products(product_id)
);

-- ============================================================
-- Table: transactions
-- Guarda o histórico de uso de cada produto (débito no saldo).
-- ============================================================
CREATE TABLE IF NOT EXISTS transactions (
    transaction_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    transaction_value DECIMAL(10, 2) NOT NULL, -- Valor debitado pela transação
    transaction_date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (product_id) REFERENCES products(product_id)
);

-- ============================================================
-- Indexes
-- ============================================================
-- Índices na tabela payments
CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments (user_id);
CREATE INDEX IF NOT EXISTS idx_payments_product_id ON payments (product_id);

-- Índices na tabela transactions
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions (user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_product_id ON transactions (product_id);

-- ============================================================
-- TRIGGERS PARA MANTER O SALDO (current_balance) EM 'users'
-- ============================================================
-- Quando inserir um novo 'payment', somar o valor em current_balance do usuário
CREATE TRIGGER IF NOT EXISTS trg_update_balance_after_payment
AFTER INSERT ON payments
BEGIN
    UPDATE users
    SET current_balance = current_balance + NEW.top_up_value
    WHERE user_id = NEW.user_id;
END;

-- Quando inserir uma nova 'transaction', subtrair o valor em current_balance do usuário
CREATE TRIGGER IF NOT EXISTS trg_update_balance_after_transaction
AFTER INSERT ON transactions
BEGIN
    UPDATE users
    SET current_balance = current_balance - NEW.transaction_value
    WHERE user_id = NEW.user_id;
END;

-- ============================================================
-- EXEMPLO DE INSERÇÕES DE TESTE
-- ============================================================

-- ------------------------------------------------------------
-- 1) Inserir usuários
--    current_balance inicia em 0, pois as recargas vão alterar isso.
-- ------------------------------------------------------------
INSERT INTO users (full_name, email, password_hash)
VALUES
  ('John Doe', 'john.doe@example.com', 'hashsenha123'),
  ('Jane Smith', 'jane.smith@example.com', 'hashsenha456');

-- ------------------------------------------------------------
-- 2) Inserir produtos
-- ------------------------------------------------------------
INSERT INTO products (product_name, description, price)
VALUES
  ('TheFunction', 'Plugin de espacialidade avançada', 0.20),
  ('TheWave', 'Plugin da ondinha', 0.20);

-- ------------------------------------------------------------
-- 3) Inserir pagamentos (recargas) - Teste do Trigger de Payment
--    - Pagamento 1: John recarrega R$50
--    - Pagamento 2: Jane recarrega R$100
--    Observação: Ao inserir cada pagamento, o trigger vai somar em current_balance do usuário.
-- ------------------------------------------------------------
INSERT INTO payments (user_id, product_id, external_payment_id, supplier_name, top_up_value, balance_after_top_up)
VALUES
  (1, NULL, 'EXT-ABC-123', 'PayPal', 50.00, 50.00),
  (2, NULL, 'EXT-XYZ-789', 'Stripe', 100.00, 100.00);

-- ------------------------------------------------------------
-- 4) Inserir transações (débito) - Teste do Trigger de Transaction
--    - John compra 'Product A' (10.00)
--    - Jane compra 'Product B' (20.00)
--    Cada transação debita 'transaction_value' do saldo via trigger.
-- ------------------------------------------------------------
INSERT INTO transactions (user_id, product_id, transaction_value)
VALUES
  (1, 1, 10.00),   -- John compra Product A (10.00)
  (2, 2, 20.00);   -- Jane compra Product B (20.00)

-- John faz mais uma compra de 'Product B' (20.00)
INSERT INTO transactions (user_id, product_id, transaction_value)
VALUES
  (1, 2, 20.00);

-- ============================================================
-- FIM DO SCRIPT
-- ============================================================

-- Para verificar o saldo final:
-- SELECT user_id, full_name, current_balance FROM users;
--
--  - John:
--     * Saldo inicial: 0
--     * Recebe 50  => 50
--     * Compra 10 => 40
--     * Compra 20 => 20
--
--  - Jane:
--     * Saldo inicial: 0
--     * Recebe 100 => 100
--     * Compra 20 => 80
