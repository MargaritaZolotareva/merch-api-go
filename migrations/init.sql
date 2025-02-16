CREATE SEQUENCE merch_id_seq INCREMENT BY 1 MINVALUE 1 START 1;
CREATE SEQUENCE purchase_id_seq INCREMENT BY 1 MINVALUE 1 START 1;
CREATE SEQUENCE transaction_id_seq INCREMENT BY 1 MINVALUE 1 START 1;
CREATE SEQUENCE employee_id_seq INCREMENT BY 1 MINVALUE 1 START 1;

CREATE TABLE merch
(
    id    SERIAL PRIMARY KEY,
    name  VARCHAR(32) NOT NULL,
    price INT NOT NULL
);

CREATE UNIQUE INDEX idx_unique_merch_name ON merch (name);

CREATE TABLE purchase
(
    id          SERIAL PRIMARY KEY,
    employee_id INT NOT NULL,
    merch_id    INT NOT NULL,
    created_at  timestamp DEFAULT now()
);

CREATE INDEX idx_purchase_employee_id ON purchase (employee_id);
CREATE INDEX idx_purchase_merch_id ON purchase (merch_id);

CREATE TABLE transaction
(
    id          SERIAL PRIMARY KEY,
    receiver_id INT NOT NULL,
    sender_id   INT NOT NULL,
    amount      INT NOT NULL,
    created_at  timestamp DEFAULT now()
);

CREATE INDEX idx_transaction_receiver_id ON transaction (receiver_id);
CREATE INDEX idx_transaction_sender_id ON transaction (sender_id);

CREATE TABLE employee
(
    id       SERIAL PRIMARY KEY,
    username VARCHAR(32)  NOT NULL,
    password VARCHAR(255) NOT NULL,
    balance  INT          NOT NULL
);

CREATE INDEX idx_unique_employee_username ON employee (username);

ALTER TABLE purchase
    ADD CONSTRAINT fk_purchase_employee_id_employee_id FOREIGN KEY (employee_id) REFERENCES employee (id) NOT DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE purchase
    ADD CONSTRAINT fk_purchase_merch_id_merch_id FOREIGN KEY (merch_id) REFERENCES merch (id) NOT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE transaction
    ADD CONSTRAINT fk_transaction_receiver_id_employee_id FOREIGN KEY (receiver_id) REFERENCES employee (id) NOT DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE transaction
    ADD CONSTRAINT fk_transaction_sender_id_employee_id FOREIGN KEY (sender_id) REFERENCES employee (id) NOT DEFERRABLE INITIALLY IMMEDIATE;

INSERT INTO merch (name, price) VALUES
                                    ('t-shirt', 80),
                                    ('cup', 20),
                                    ('book', 50),
                                    ('pen', 10),
                                    ('powerbank', 200),
                                    ('hoody', 300),
                                    ('umbrella', 200),
                                    ('socks', 10),
                                    ('wallet', 50),
                                    ('pink-hoody', 500);