-- +goose Up
-- +goose StatementBegin
CREATE TABLE otp_requests (
    id         CHAR(26)     NOT NULL,
    email      VARCHAR(320) NOT NULL,
    -- Store bcrypt hash of the 6-digit code — never the plain code
    code_hash  TEXT         NOT NULL,
    used       BOOLEAN      NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMPTZ  NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_otp_requests PRIMARY KEY (id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_otp_requests_email_active ON otp_requests (email, expires_at)
    WHERE used = FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS otp_requests;
-- +goose StatementEnd
