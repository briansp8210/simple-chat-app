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
    ts timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data_type message_data_t NOT NULL,
    contents varchar(512) NOT NULL
);

INSERT INTO users (username, password) VALUES ('a', decode('697f2d856172cb8309d6b8b97dac4de344b549d4dee61edfb4962d8698b7fa803f4f93ff24393586e28b5b957ac3d1d369420ce53332712f997bd336d09ab02a', 'hex'));
INSERT INTO users (username, password) VALUES ('b', decode('8446c46ee03793ba6e5813ba0db4480008926dd1d19efe2c8eb92f9034da974d2171ae483f29ce3a79ed4fdd621ae1ed14fe12532af95ddd0728779ce5aa842d', 'hex'));
INSERT INTO users (username, password) VALUES ('c', decode('bfe4d7f7377116dc15f794d902621797b72b32396382de2b6e49d4f1d7eabdfddcfc3bc127bb67f92f9458a5733bb21804e7ccd56b4b6f81049339f477cd279d', 'hex'));
INSERT INTO conversations (name, type, member_ids) VALUES ('A->B', 'PRIVATE', '{1, 2}');
INSERT INTO conversations (name, type, member_ids) VALUES ('A->C', 'PRIVATE', '{1, 3}');
INSERT INTO messages (sender_id, conversation_id, data_type, contents) VALUES (1, 1, 'TEXT', 'Hello B, I am A');
INSERT INTO messages (sender_id, conversation_id, data_type, contents) VALUES (1, 1, 'TEXT', 'This is a direct message');
INSERT INTO messages (sender_id, conversation_id, data_type, contents) VALUES (2, 1, 'TEXT', 'Hi~ message received');
INSERT INTO messages (sender_id, conversation_id, data_type, contents) VALUES (1, 1, 'TEXT', 'Cool');
INSERT INTO messages (sender_id, conversation_id, data_type, contents) VALUES (1, 2, 'TEXT', 'Hi! C, I am A');
