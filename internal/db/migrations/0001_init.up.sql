CREATE TABLE users (
    uuid UUID PRIMARY KEY,
    login VARCHAR(255) UNIQUE NOT NULL,
    hashed_password TEXT NOT NULL,
    balance BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE orders (
    id BIGINT PRIMARY KEY,
    user_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    accrual int null CHECK (accrual > 0),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    CONSTRAINT fk__orders__user
        FOREIGN KEY (user_id) 
        REFERENCES users(uuid)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT
);

CREATE INDEX idx__orders__user_id ON orders(user_id);
CREATE INDEX idx__orders__status ON orders(status);

CREATE TABLE transactions (
    uuid UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    order_id BIGINT NOT NULL,
    type VARCHAR(20) NOT NULL,
    amount INT NOT NULL CHECK (amount > 0),
    created_at TIMESTAMP NOT NULL,
    CONSTRAINT fk__transactions__user
        FOREIGN KEY (user_id) 
        REFERENCES users(uuid)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT
);

CREATE INDEX idx__transactions__user_id ON transactions(user_id);
CREATE INDEX idx__transactions__order_id ON transactions(order_id);
