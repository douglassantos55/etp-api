CREATE TABLE IF NOT EXISTS `loans` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `company_id` INTEGER NOT NULL,
    `principal` INTEGER NOT NULL,
    `interest_rate` DECIMAL(4, 2) NOT NULL,
    `payable_from` TIMESTAMP NOT NULL,
    `interest_paid` INTEGER DEFAULT 0,
    `principal_paid` INTEGER DEFAULT 0,
    `delayed_payments` TINYINT DEFAULT 0,
    FOREIGN KEY (`company_id`) REFERENCES `companies`(`id`)
);
