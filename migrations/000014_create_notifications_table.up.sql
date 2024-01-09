CREATE TABLE IF NOT EXISTS `notifications` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `message` TEXT,
    `company_id` INTEGER,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`company_id`) REFERENCES `companies` (`id`)
);
