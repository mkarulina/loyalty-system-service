-- Таблица пользователей --
CREATE TABLE IF NOT EXISTS users (
    token VARCHAR(255),
    login VARCHAR(255) UNIQUE,
    password VARCHAR(255)
                                 );

-- Таблица заказов --
CREATE TABLE IF NOT EXISTS orders (
    user_id VARCHAR(255),
    number VARCHAR(255) UNIQUE,
    status VARCHAR(255) DEFAULT 'NEW',
    accrual FLOAT DEFAULT 0,
    withdrawn FLOAT DEFAULT 0,
    uploaded_at TIMESTAMP
                                  );

-- Таблица истории списаний --
CREATE TABLE IF NOT EXISTS withdrawals_history (
    user_id VARCHAR(255),
    order_number VARCHAR(255),
    sum FLOAT DEFAULT 0,
    uploaded_at TIMESTAMP
                                               );