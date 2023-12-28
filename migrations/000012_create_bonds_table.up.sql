CREATE TABLE IF NOT EXISTS `bonds` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `company_id` INTEGER NOT NULL,
    `amount` INTEGER NOT NULL,
    `interest_rate` DECIMAL(4, 2) NOT NULL,
    FOREIGN KEY (`company_id`) REFERENCES `companies`(`id`)
);

CREATE TABLE IF NOT EXISTS `bonds_creditors` (
    `company_id` INTEGER NOT NULL,
    `bond_id` INTEGER NOT NULL,
    `interest_rate` DECIMAL(4, 2) NOT NULL,
    `interest_paid` INTEGER DEFAULT 0,
    `payable_from` TIMESTAMP NOT NULL,
    `principal` INTEGER NOT NULL,
    `principal_paid` INTEGER DEFAULT 0,
    `delayed_payments` TINYINT DEFAULT 0,
    FOREIGN KEY (`company_id`) REFERENCES `companies`(`id`),
    FOREIGN KEY (`bond_id`) REFERENCES `bonds`(`id`),
    PRIMARY KEY (`company_id`, `bond_id`)
);
