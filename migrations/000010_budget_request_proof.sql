-- 000010_budget_request_proof.sql
-- Add proof_url for budget request creation, and approval_notes + approval_proof_url for approve/reject

ALTER TABLE budget_requests
    ADD COLUMN proof_url VARCHAR(500) AFTER reason,
    ADD COLUMN approval_notes TEXT AFTER approved_by,
    ADD COLUMN approval_proof_url VARCHAR(500) AFTER approval_notes;
