-- Add proof_url column to expense_approvals for storing bukti transfer
ALTER TABLE expense_approvals ADD COLUMN proof_url VARCHAR(500) DEFAULT '' AFTER notes;
