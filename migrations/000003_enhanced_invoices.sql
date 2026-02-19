-- 000003_enhanced_invoices.sql
-- Enhanced invoice system: invoice types, line items, company settings, approve/reject

-- ============================================
-- 1. ALTER invoices table - add new columns
-- ============================================
ALTER TABLE invoices
    ADD COLUMN invoice_type ENUM('DP','FINAL_PAYMENT','TOP_1','TOP_2','TOP_3','MEALS','ADDITIONAL') NOT NULL DEFAULT 'DP' AFTER invoice_number,
    ADD COLUMN status ENUM('PENDING','APPROVED','REJECTED') NOT NULL DEFAULT 'PENDING' AFTER amount,
    ADD COLUMN recipient_name VARCHAR(255) NOT NULL DEFAULT '' AFTER file_url,
    ADD COLUMN recipient_address TEXT AFTER recipient_name,
    ADD COLUMN attention VARCHAR(255) DEFAULT NULL AFTER recipient_address,
    ADD COLUMN po_number VARCHAR(100) DEFAULT NULL AFTER attention,
    ADD COLUMN invoice_date DATE NOT NULL DEFAULT (CURRENT_DATE) AFTER po_number,
    ADD COLUMN dp_percentage DECIMAL(5,2) DEFAULT NULL AFTER invoice_date,
    ADD COLUMN subtotal DECIMAL(18,2) NOT NULL DEFAULT 0.00 AFTER dp_percentage,
    ADD COLUMN tax_percentage DECIMAL(5,2) NOT NULL DEFAULT 0.00 AFTER subtotal,
    ADD COLUMN tax_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00 AFTER tax_percentage,
    ADD COLUMN notes TEXT AFTER tax_amount,
    ADD COLUMN language ENUM('ID','EN') NOT NULL DEFAULT 'ID' AFTER notes,
    ADD COLUMN approved_by BIGINT UNSIGNED DEFAULT NULL AFTER created_by,
    ADD COLUMN reject_notes TEXT AFTER approved_by,
    ADD INDEX idx_invoices_status (status),
    ADD INDEX idx_invoices_type (invoice_type),
    ADD CONSTRAINT fk_inv_approved_by FOREIGN KEY (approved_by) REFERENCES users(id);

-- Make file_url optional (not all invoices need uploaded file)
ALTER TABLE invoices MODIFY COLUMN file_url VARCHAR(500) DEFAULT NULL;

-- ============================================
-- 2. CREATE invoice_items table
-- ============================================
CREATE TABLE IF NOT EXISTS invoice_items (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    invoice_id BIGINT UNSIGNED NOT NULL,
    description VARCHAR(500) NOT NULL,
    quantity DECIMAL(10,2) NOT NULL DEFAULT 1.00,
    unit VARCHAR(50) NOT NULL DEFAULT 'unit',
    unit_price DECIMAL(18,2) NOT NULL,
    subtotal DECIMAL(18,2) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_invoice_items_invoice (invoice_id),
    CONSTRAINT fk_ii_invoice FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- 3. CREATE company_settings table
-- ============================================
CREATE TABLE IF NOT EXISTS company_settings (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    company_name VARCHAR(255) NOT NULL,
    company_code VARCHAR(10) NOT NULL,
    address TEXT,
    phone VARCHAR(50),
    email VARCHAR(255),
    npwp VARCHAR(50),
    bank_name VARCHAR(100),
    bank_account_number VARCHAR(50),
    bank_account_name VARCHAR(255),
    bank_branch VARCHAR(255),
    logo_url VARCHAR(500),
    signatory_name VARCHAR(255),
    signatory_title VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- 4. Add invoice notification types
-- ============================================
-- (notification types are defined in Go code, no DB change needed)

-- ============================================
-- 5. Update invoice_number format
-- ============================================
-- New format: {seq}/INV/{company_code}/{MM}/{YYYY}
-- Example: 001/INV/CGA/02/2026
-- The format change is handled in Go code, existing data stays as-is
