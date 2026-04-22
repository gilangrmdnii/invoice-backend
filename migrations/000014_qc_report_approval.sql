-- Approval fields for QC Reports (handled by QC_COORDINATOR)
ALTER TABLE qc_reports
    ADD COLUMN status ENUM('DRAFT','PENDING','APPROVED','REJECTED') NOT NULL DEFAULT 'DRAFT' AFTER total_amount,
    ADD COLUMN approved_by BIGINT UNSIGNED NULL AFTER status,
    ADD COLUMN approval_notes TEXT NULL AFTER approved_by,
    ADD COLUMN approved_at TIMESTAMP NULL AFTER approval_notes,
    ADD INDEX idx_qc_reports_status (status),
    ADD FOREIGN KEY (approved_by) REFERENCES users(id);
