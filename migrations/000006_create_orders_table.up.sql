CREATE TABLE IF NOT EXISTS `orders` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `quantity` INTEGER UNSIGNED NOT NULL,
    `quality` TINYINT UNSIGNED NOT NULL,
    `price` INTEGER UNSIGNED NOT NULL,
    `sourcing_cost` INTEGER UNSIGNED DEFAULT 0,
    `market_fee` INTEGER UNSIGNED DEFAULT 0,
    `transport_fee` INTEGER UNSIGNED DEFAULT 0,
    `company_id` INTEGER NOT NULL,
    `resource_id` INTEGER NOT NULL,
    `purchased_at` TIMESTAMP DEFAULT NULL,
    `canceled_at` TIMESTAMP DEFAULT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`company_id`) REFERENCES `companies`(`id`),
    FOREIGN KEY (`resource_id`) REFERENCES `resources`(`id`)
);
