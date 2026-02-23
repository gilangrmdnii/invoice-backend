-- 000005_invoice_payments.sql
-- Payment tracking for invoices

-- 1. Add payment_status and due_date to invoices
ALTER TABLE invoices
    ADD COLUMN payment_status ENUM('UNPAID','PARTIAL_PAID','PAID') NOT NULL DEFAULT 'UNPAID' AFTER status,
    ADD COLUMN due_date DATE DEFAULT NULL AFTER invoice_date,
    ADD COLUMN paid_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00 AFTER amount;

-- 2. Create invoice_payments table
CREATE TABLE IF NOT EXISTS invoice_payments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    invoice_id BIGINT UNSIGNED NOT NULL,
    amount DECIMAL(18,2) NOT NULL,
    payment_date DATE NOT NULL,
    payment_method ENUM('TRANSFER','CASH','GIRO','OTHER') NOT NULL DEFAULT 'TRANSFER',
    proof_url VARCHAR(500) DEFAULT NULL,
    notes TEXT DEFAULT NULL,
    created_by BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_ip_invoice (invoice_id),
    CONSTRAINT fk_ip_invoice FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE,
    CONSTRAINT fk_ip_created_by FOREIGN KEY (created_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
