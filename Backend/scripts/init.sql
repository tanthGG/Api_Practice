CREATE DATABASE IF NOT EXISTS loans_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE loans_db;

CREATE TABLE IF NOT EXISTS loan_applications (
    id CHAR(36) NOT NULL PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL,
    monthly_income DECIMAL(12,2) NOT NULL,
    loan_amount DECIMAL(12,2) NOT NULL,
    loan_purpose VARCHAR(50) NOT NULL,
    age INT NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    email VARCHAR(255) NOT NULL,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    unix_ts BIGINT NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
