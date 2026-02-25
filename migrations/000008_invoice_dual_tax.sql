-- Rename existing tax columns and add PPh columns for dual tax system
ALTER TABLE invoices
  CHANGE COLUMN tax_percentage ppn_percentage DECIMAL(5,2) NOT NULL DEFAULT 0,
  CHANGE COLUMN tax_amount ppn_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
  ADD COLUMN pph_percentage DECIMAL(5,2) NOT NULL DEFAULT 0 AFTER ppn_amount,
  ADD COLUMN pph_amount DECIMAL(15,2) NOT NULL DEFAULT 0 AFTER pph_percentage;
