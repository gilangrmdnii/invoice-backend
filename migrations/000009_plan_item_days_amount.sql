-- Add days (hari) and amount (jumlah) columns to project_plan_items
ALTER TABLE project_plan_items
  ADD COLUMN days INT NOT NULL DEFAULT 0 AFTER unit_price,
  ADD COLUMN amount DECIMAL(18,2) NOT NULL DEFAULT 0 AFTER days;
