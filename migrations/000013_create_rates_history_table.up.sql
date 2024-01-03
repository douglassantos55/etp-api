CREATE TABLE IF NOT EXISTS `rates_history` (
    `period` DATE PRIMARY KEY,
    `inflation` DECIMAL(5, 4),
    `interest` DECIMAL(5, 4)
);
