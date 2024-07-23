-- +goose Up
-- +goose StatementBegin
BEGIN;

ALTER TABLE orders ADD COLUMN weight FLOAT DEFAULT 0.0;
ALTER TABLE orders ADD COLUMN cost FLOAT DEFAULT 0.0;
ALTER TABLE orders ADD COLUMN package_type TEXT DEFAULT 'film';
ALTER TABLE orders ADD COLUMN package_cost FLOAT DEFAULT 1.0;

UPDATE orders SET weight = 0.0, cost = 0.0, package_type = 'film', package_cost = 1.0 WHERE weight IS NULL OR cost IS NULL OR package_type IS NULL OR package_cost IS NULL;

ALTER TABLE orders ALTER COLUMN weight SET NOT NULL;
ALTER TABLE orders ALTER COLUMN cost SET NOT NULL;
ALTER TABLE orders ALTER COLUMN package_type SET NOT NULL;
ALTER TABLE orders ALTER COLUMN package_cost SET NOT NULL;

COMMIT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE orders DROP COLUMN weight;
ALTER TABLE orders DROP COLUMN cost;
ALTER TABLE orders DROP COLUMN package_type;
ALTER TABLE orders DROP COLUMN package_cost;
-- +goose StatementEnd
