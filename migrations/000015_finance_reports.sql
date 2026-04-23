-- Finance Report: Recruiter fees + Sample entries (per project)
-- Aggregation pengeluaran per member/category diambil dari table expenses (existing)

-- Perolehan Recruit: Fee recruiter per project
CREATE TABLE IF NOT EXISTS finance_recruiter_fees (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    project_id BIGINT UNSIGNED NOT NULL,
    recruiter_name VARCHAR(255) NOT NULL,
    jumlah INT NOT NULL DEFAULT 0,
    fee_recruiter DECIMAL(15,2) NOT NULL DEFAULT 0,
    insentif_responden_main DECIMAL(15,2) NOT NULL DEFAULT 0,
    jumlah_responden_main INT NOT NULL DEFAULT 0,
    insentif_responden_backup DECIMAL(15,2) NOT NULL DEFAULT 0,
    jumlah_responden_backup INT NOT NULL DEFAULT 0,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    INDEX idx_finance_recruiter_project (project_id)
);

-- Tabel Sample per Tanggal Pelaksanaan
CREATE TABLE IF NOT EXISTS finance_sample_entries (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    project_id BIGINT UNSIGNED NOT NULL,
    tanggal_pelaksanaan DATE NOT NULL,
    jumlah_sample INT NOT NULL DEFAULT 0,
    insentif_responden_main DECIMAL(15,2) NOT NULL DEFAULT 0,
    jumlah_responden_main INT NOT NULL DEFAULT 0,
    insentif_responden_backup DECIMAL(15,2) NOT NULL DEFAULT 0,
    jumlah_responden_backup INT NOT NULL DEFAULT 0,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    INDEX idx_finance_sample_project (project_id)
);

-- Manual expense entries untuk kategori yang tidak ada di expenses table
-- (dipakai Finance kalau mau input langsung di laporan, bukan via SPV)
CREATE TABLE IF NOT EXISTS finance_manual_expenses (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    project_id BIGINT UNSIGNED NOT NULL,
    member_user_id BIGINT UNSIGNED NULL,
    member_name VARCHAR(255) NOT NULL DEFAULT '',
    category VARCHAR(100) NOT NULL,
    tanggal DATE NULL,
    description VARCHAR(500) NOT NULL DEFAULT '',
    quantity INT NOT NULL DEFAULT 1,
    unit_price DECIMAL(15,2) NOT NULL DEFAULT 0,
    amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    sort_order INT NOT NULL DEFAULT 0,
    created_by BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (member_user_id) REFERENCES users(id),
    FOREIGN KEY (created_by) REFERENCES users(id),
    INDEX idx_finance_manual_expenses_project (project_id)
);
