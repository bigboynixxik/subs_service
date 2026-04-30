-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS subscriptions
(
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_name TEXT    NOT NULL,
    price        INTEGER NOT NULL,
    user_id      UUID    NOT NULL,
    start_date   DATE    NOT NULL,
    end_date     DATE
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS subscriptions;
-- +goose StatementEnd