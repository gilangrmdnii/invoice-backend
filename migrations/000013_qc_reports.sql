-- QC Financial Report (Laporan Keuangan QC per Project)
CREATE TABLE IF NOT EXISTS qc_reports (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    project_id BIGINT UNSIGNED NOT NULL,
    qc_user_id BIGINT UNSIGNED NOT NULL,
    spv_names VARCHAR(500) NOT NULL DEFAULT '',
    project_type ENUM('KUALITATIF','KUANTITATIF') NOT NULL DEFAULT 'KUALITATIF',
    methodology ENUM('FGD_TRIAD','HOME_VISIT','CLT','IDI','RANDOM') NOT NULL DEFAULT 'FGD_TRIAD',
    city VARCHAR(255) NOT NULL DEFAULT '',
    area ENUM('URBAN','RURAL','URBAN_RURAL') NOT NULL DEFAULT 'URBAN',
    execution_start_date DATE NULL,
    execution_end_date DATE NULL,
    briefing_date DATE NULL,
    work_start_date DATE NULL,
    work_end_date DATE NULL,
    visit_target INT NOT NULL DEFAULT 0,
    visit_ok INT NOT NULL DEFAULT 0,
    telp_target INT NOT NULL DEFAULT 0,
    telp_ok INT NOT NULL DEFAULT 0,
    total_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    location VARCHAR(255) NOT NULL DEFAULT '',
    report_date DATE NULL,
    qc_signatory_name VARCHAR(255) NOT NULL DEFAULT '',
    qc_signatory_title VARCHAR(255) NOT NULL DEFAULT 'Quality Control',
    coordinator_signatory_name VARCHAR(255) NOT NULL DEFAULT '',
    coordinator_signatory_title VARCHAR(255) NOT NULL DEFAULT 'Koordinator QC',
    note TEXT,
    created_by BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (qc_user_id) REFERENCES users(id),
    FOREIGN KEY (created_by) REFERENCES users(id),
    INDEX idx_qc_reports_project (project_id),
    INDEX idx_qc_reports_qc_user (qc_user_id)
);

-- Line items untuk perhitungan biaya (Visit Urban/Rural OK/DO, Telp, Recording, dll)
CREATE TABLE IF NOT EXISTS qc_report_items (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    qc_report_id BIGINT UNSIGNED NOT NULL,
    category ENUM(
        'VISIT_URBAN','VISIT_RURAL','TELP_QUAL','TELP_QUANT','CLT_TIMESHEET',
        'RECORDING','UANG_MAKAN','INPUT_PERPI','PARKIR','BENSIN','LAIN_LAIN'
    ) NOT NULL,
    status ENUM('OK','DO','NONE') NOT NULL DEFAULT 'NONE',
    label VARCHAR(255) NOT NULL DEFAULT '',
    quantity INT NOT NULL DEFAULT 0,
    unit_price DECIMAL(15,2) NOT NULL DEFAULT 0,
    subtotal DECIMAL(15,2) NOT NULL DEFAULT 0,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (qc_report_id) REFERENCES qc_reports(id) ON DELETE CASCADE,
    INDEX idx_qc_report_items_report (qc_report_id)
);

-- Recruiter performance (tabel perolehan recruiter per QC report)
CREATE TABLE IF NOT EXISTS qc_recruiter_performance (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    qc_report_id BIGINT UNSIGNED NOT NULL,
    recruiter_name VARCHAR(255) NOT NULL,
    total INT NOT NULL DEFAULT 0,
    ok_perpi INT NOT NULL DEFAULT 0,
    do_perpi INT NOT NULL DEFAULT 0,
    ok_qc INT NOT NULL DEFAULT 0,
    do_qc INT NOT NULL DEFAULT 0,
    notes TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (qc_report_id) REFERENCES qc_reports(id) ON DELETE CASCADE,
    INDEX idx_qc_recruiter_report (qc_report_id)
);
