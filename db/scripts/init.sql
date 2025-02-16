CREATE TABLE IF NOT EXISTS Users (
    username VARCHAR(50) PRIMARY KEY,
    password VARCHAR(100) NOT NULL,
    coins INT DEFAULT 0
    );

CREATE TABLE IF NOT EXISTS Inventory (
    type VARCHAR(30) PRIMARY KEY,
    quantity INT DEFAULT 0,
    owner VARCHAR(50) NOT NULL REFERENCES Users(username)
    );

CREATE TABLE IF NOT EXISTS History (
    "from" VARCHAR(50) NOT NULL,
    "to" VARCHAR(50) NOT NULL,
    amount INT NOT NULL
    )