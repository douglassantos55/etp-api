CREATE TABLE IF NOT EXISTS `classifications` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `name` VARCHAR(255) NOT NULL,
    `parent_id` INTEGER DEFAULT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP DEFAULT NULL,
    FOREIGN KEY (`parent_id`) REFERENCES `classifications` (`id`)
);

CREATE TABLE IF NOT EXISTS `transactions` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `value` INTEGER,
    `company_id` INTEGER,
    `description` VARCHAR(255),
    `classification_id` INTEGER,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`company_id`) REFERENCES `companies` (`id`),
    FOREIGN KEY (`classification_id`) REFERENCES `classifications` (`id`)
);
