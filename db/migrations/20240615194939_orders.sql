-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS orders (
    order_id INT PRIMARY KEY,
    pick_point_id INT NOT NULL,
    client_id INT NOT NULL,
    added_date TIMESTAMP NOT NULL,
    shelf_life TIMESTAMP NOT NULL,
    issued BOOLEAN NOT NULL DEFAULT FALSE,
    issue_date TIMESTAMP DEFAULT '0001-01-01 00:00:00',
    returned BOOLEAN NOT NULL DEFAULT FALSE,
    return_date TIMESTAMP DEFAULT '0001-01-01 00:00:00',
    deleted BOOLEAN NOT NULL DEFAULT FALSE,
    delete_date TIMESTAMP DEFAULT '0001-01-01 00:00:00',
    order_hash TEXT UNIQUE NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd