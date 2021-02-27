CREATE TABLE users (
    id serial PRIMARY KEY,
    username varchar(32) NOT NULL UNIQUE,
    password bytea NOT NULL
);

CREATE TYPE conversation_t AS ENUM ('PRIVATE');
CREATE TYPE message_data_t AS ENUM ('TEXT');

CREATE TABLE conversations (
    id serial PRIMARY KEY,
    name varchar(32) NOT NULL,
    type conversation_t NOT NULL,
    member_ids integer[] NOT NULL DEFAULT '{}'::integer[]
);

CREATE TABLE messages (
    id serial PRIMARY KEY,
    sender_id integer REFERENCES users NOT NULL,
    conversation_id integer REFERENCES conversations NOT NULL,
    ts timestamp NOT NULL,
    data_type message_data_t NOT NULL,
    contents varchar(512) NOT NULL
);
