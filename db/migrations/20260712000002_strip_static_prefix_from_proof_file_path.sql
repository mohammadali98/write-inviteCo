-- migrate:up
UPDATE order_payments
SET proof_file_path = regexp_replace(proof_file_path, '^/static/payment-proofs/', '')
WHERE proof_file_path LIKE '/static/payment-proofs/%';

-- migrate:down
UPDATE order_payments
SET proof_file_path = '/static/payment-proofs/' || proof_file_path
WHERE proof_file_path IS NOT NULL
  AND proof_file_path NOT LIKE '/static/payment-proofs/%'
  AND proof_file_path NOT LIKE '%/%';
