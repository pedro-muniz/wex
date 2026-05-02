CREATE TABLE purchase_transactions (
    id UUID PRIMARY KEY,
    description VARCHAR(50) NOT NULL,
    transaction_date TIMESTAMP NOT NULL,
    amount DECIMAL(19, 4) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
