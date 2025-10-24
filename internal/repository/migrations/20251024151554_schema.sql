-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS wallets (
    id TEXT PRIMARY KEY, -- выбрал TEXT на случай, если впоследствии изменится тип идентификатора
    amount INTEGER NOT NULL CHECK (amount >= 0), -- сумма в минимальных единицах валюты (копейки, центы и т.д.)
                                                 -- благодаря проверке смогу различать, существует ли кошелек или на нем не хватает средств
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- минимальная дополнительная информация о времени создания
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS transactions ( -- таблица для хранения данных о поступлениях и списаниях
    id TEXT PRIMARY KEY, 
    wallet_id TEXT NOT NULL REFERENCES wallets(id) ON DELETE CASCADE, -- внешний ключ на кошелек
    amount INTEGER NOT NULL,
    operation_type TEXT NOT NULL, 
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP -- минимальная дополнительная информация о времени создания
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE transactions;
DROP TABLE wallets;
-- +goose StatementEnd
