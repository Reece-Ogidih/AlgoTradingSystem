CREATE DATABASE IF NOT EXISTS ct_bot;

USE ct_bot;
CREATE TABLE IF NOT EXISTS candles_1m (
    id INT NOT NULL AUTO_INCREMENT,
    open_times_ms BIGINT,
    open DOUBLE,
    close DOUBLE, 
    high DOUBLE,
    low DOUBLE,
    volume DOUBLE,
    is_final BOOLEAN,
    PRIMARY KEY (id)
);