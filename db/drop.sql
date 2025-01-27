-- Desabilitar temporariamente checks de foreign_key (opcional)
PRAGMA foreign_keys = OFF;

-- Dropar triggers
DROP TRIGGER IF EXISTS trg_update_balance_after_payment;
DROP TRIGGER IF EXISTS trg_update_balance_after_transaction;

-- Dropar tabelas
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS users;

-- Reativar checks de foreign_key
PRAGMA foreign_keys = ON;
