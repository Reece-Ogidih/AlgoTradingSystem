CREATE DATABASE IF NOT EXISTS ct_bot;

USE ct_bot;

DROP TABLE IF EXISTS candles_1m; -- A previous table I made which is unneccesary now

-- Historical 1m candles
CREATE TABLE IF NOT EXISTS hist_candles_1m (
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

-- Also am thinking ahead that to be able to assess Bot performace it would be useful to have the bot log all trades into a table
-- Would thus also be worthwhile to save the live candle data to cross reference with the trades documented

-- Live 1m closing prices 
CREATE TABLE IF NOT EXISTS live_candles_1m (
    id INT NOT NULL AUTO_INCREMENT,
    open_times_ms BIGINT NOT NULL UNIQUE,   -- Save the opentime of the candle (unique so no duplicates)
    close DOUBLE,                           -- Save the close price
    is_final BOOLEAN,
    PRIMARY KEY (id)
);

-- Trade logs table
CREATE TABLE IF NOT EXISTS bot_trades (
    id INT NOT NULL AUTO_INCREMENT,
    timestamp_ms BIGINT NOT NULL,           -- To cross reference with the opentime
    action ENUM('BUY', 'SELL') NOT NULL,    -- Need to document the action of sell or buy
    price DOUBLE NOT NULL,                  -- The price at which the action was held
    quantity DOUBLE,                        -- The quantity which allows to infer total investment
    confidence_score DOUBLE,                -- The output from the ML algorithm
    notes TEXT,                             -- For now I named this notes but will likely record ADX val and anything else important to making this decision
    PRIMARY KEY (id)
);

SHOW TABLES;

-- To train ML model, will be good to add to the historical data both the ADX as well as a variable to depict when the bot would BUY or SELL
CREATE TABLE IF NOT EXISTS train_ml (
    id INT NOT NULL AUTO_INCREMENT,
    open_times_ms BIGINT,
    open DOUBLE,
    close DOUBLE, 
    high DOUBLE,
    low DOUBLE,
    volume DOUBLE,
    adx BOOLEAN,
    signal TINYINT, -- Will encode -1 for sell order, 0 for hold and 1 for buy order. Note these are catageroical not ordered.
    PRIMARY KEY (id)
);
