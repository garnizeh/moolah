-- +goose Up
CREATE TYPE asset_type AS ENUM (
  'stock', 'bond', 'fund', 'crypto', 'real_estate', 'income_source'
);
CREATE TYPE income_type AS ENUM (
  'none', 'dividend', 'coupon', 'rent', 'interest', 'salary'
);
CREATE TYPE receivable_status AS ENUM ('pending', 'received', 'cancelled');

-- +goose Down
DROP TYPE IF EXISTS receivable_status;
DROP TYPE IF EXISTS income_type;
DROP TYPE IF EXISTS asset_type;
