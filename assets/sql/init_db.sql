CREATE TABLE IF NOT EXISTS bank (
    id            INTEGER  NOT NULL PRIMARY KEY UNIQUE,
    currency_code TEXT     NOT NULL,
    month_id      INTEGER  NOT NULL,
    db_version    TEXT
) STRICT;


CREATE TABLE IF NOT EXISTS months (
    id    INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    -- If my program survives to the year 3000, then something went wrong lol
    date TEXT
)STRICT;


CREATE TABLE IF NOT EXISTS income (
    id     INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name   TEXT    NOT NULL UNIQUE,
    amount INTEGER NOT NULL CHECK (amount>0),
    period TEXT    NOT NULL CHECK (period IN ('yearly', 'monthly', 'biweekly', 'weekly'))
)STRICT;


CREATE TABLE IF NOT EXISTS income_history (
    id        INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    income_id INTEGER NOT NULL,
    month_id  INTEGER NOT NULL,
    amount    INTEGER CHECK (amount>0),
    FOREIGN KEY (income_id) REFERENCES income (id),
    FOREIGN KEY (month_id) REFERENCES months (id)
)STRICT;


-- When income has been raised through bonuses, overtime, etc..
CREATE TABLE IF NOT EXISTS income_affixes (
    id         INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    history_id INTEGER NOT NULL,
    name       TEXT    NOT NULL,
    amount     INTEGER CHECK (amount>0),
    FOREIGN KEY (history_id) REFERENCES income_history (id)
)STRICT;


CREATE TABLE IF NOT EXISTS bank_accounts (
    id       INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name     TEXT    NOT NULL,
    -- Should only store the encrypted value
    acct_num TEXT,
    -- Should only store the encrypted value
    notes    TEXT
)STRICT;


CREATE TABLE IF NOT EXISTS bank_account_history (
    id         INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL,
    month_id   INTEGER NOT NULL,
    balance    INTEGER DEFAULT 0,
    FOREIGN KEY (month_id) REFERENCES months (id),
    FOREIGN KEY (account_id) REFERENCES bank_accounts (id)
)STRICT;


CREATE TABLE IF NOT EXISTS bank_transfers (
    id              INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    bank_history_id INTEGER NOT NULL,
    month_id        INTEGER NOT NULL,
    name            TEXT    NOT NULL,
    amount          INTEGER NOT NULL,
    due_day         INTEGER NOT NULL CHECK (due_day > 0 AND due_day < 32),
    type            TEXT    NOT NULL CHECK (type IN ('withdrawal', 'deposit', 'move')),
    to_whom    TEXT,
    from_whom  TEXT,
    FOREIGN KEY (bank_history_id) REFERENCES bank_account_history (id),
    FOREIGN KEY (month_id) REFERENCES months (id)
)STRICT;


CREATE TABLE IF NOT EXISTS credit_cards (
    id               INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name             TEXT    NOT NULL UNIQUE,
    due_day          INTEGER NOT NULL CHECK (due_day > 0 AND due_day < 32),
    credit_limit     INTEGER,
    -- Should only store the encrypted value
    card_number      TEXT,
    last_four_digits TEXT   NOT NULL,
    -- Should only store the encrypted value
    notes            TEXT
)STRICT;


CREATE TABLE IF NOT EXISTS credit_card_history (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    card_id      INTEGER NOT NULL,
    month_id     INTEGER NOT NULL,
    balance      INTEGER NOT NULL,
    due_day      INTEGER NOT NULL CHECK (due_day > 0 AND due_day < 32),
    credit_limit INTEGER,
    paid_day     INTEGER CHECK (paid_day > 0 AND paid_day < 32),
    paid_amount  INTEGER,
    cleared_day  INTEGER,
    FOREIGN KEY (card_id) REFERENCES credit_cards (id),
    FOREIGN KEY (month_id) REFERENCES months (id)
)STRICT;


CREATE TABLE IF NOT EXISTS bills (
    id       INTEGER  PRIMARY KEY AUTOINCREMENT,
    type_id  INTEGER  NOT NULL,
    name     TEXT     NOT NULL UNIQUE,
    amount   INTEGER  NOT NULL,
    due_date TEXT     NOT NULL,
    status   TEXT     NOT NULL CHECK (status IN ('pending', 'missed', 'paid')),
    period   TEXT     NOT NULL CHECK (period IN ('yearly', 'monthly')),
    FOREIGN KEY (type_id) REFERENCES bill_types (id)
)STRICT;


CREATE TABLE IF NOT EXISTS bills_history (
    id          INTEGER  PRIMARY KEY AUTOINCREMENT,
    month_id    INTEGER  NOT NULL,
    type_id     INTEGER,
    name        TEXT     NOT NULL,
    amount      INTEGER  NOT NULL,
    due_date    TEXT     NOT NULL,
    paid_amount INTEGER,
    paid_date   TEXT,
    paid_how    TEXT,
    cleared_day INTEGER,
    notes       TEXT,
    FOREIGN KEY (month_id) REFERENCES months (id)
    FOREIGN KEY (type_id) REFERENCES bill_types (id)
)STRICT;


CREATE TABLE IF NOT EXISTS bills_monthly (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    bill_id   INTEGER NOT NULL UNIQUE,
    is_active INTEGER NOT NULL,
    FOREIGN KEY (bill_id) REFERENCES bills (id)
)STRICT;


CREATE TABLE IF NOT EXISTS bill_types (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT    NOT NULL UNIQUE
)STRICT;
