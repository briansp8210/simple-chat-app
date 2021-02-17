CREATE TABLE users (
    id serial PRIMARY KEY,
    username varchar(32) NOT NULL UNIQUE,
    password_hash bytea NOT NULL
);
