-- 000004_invoice_item_labels.sql
-- Add parent-child relationship to invoice_items for label grouping.
-- Items with parent_id = NULL and is_label = TRUE are section headers (labels).
-- Items with parent_id pointing to a label are child items under that label.
-- Items with parent_id = NULL and is_label = FALSE are standalone items (backward-compatible).

ALTER TABLE invoice_items
    ADD COLUMN parent_id BIGINT UNSIGNED DEFAULT NULL AFTER invoice_id,
    ADD COLUMN is_label BOOLEAN NOT NULL DEFAULT FALSE AFTER parent_id,
    ADD INDEX idx_invoice_items_parent (parent_id),
    ADD CONSTRAINT fk_ii_parent FOREIGN KEY (parent_id) REFERENCES invoice_items(id) ON DELETE CASCADE;
