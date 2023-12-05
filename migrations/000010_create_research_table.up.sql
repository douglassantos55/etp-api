CREATE TABLE IF NOT EXISTS `research_staff` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `name` VARCHAR(255) NOT NULL,
    `salary` INTEGER UNSIGNED DEFAULT 0,
    `skill` TINYINT UNSIGNED DEFAULT 0,
    `talent` TINYINT UNSIGNED DEFAULT 0,
    `status` TINYINT UNSIGNED DEFAULT 0,
    `offer` INTEGER UNSIGNED DEFAULT 0,
    `company_id` INTEGER NOT NULL,
    `poacher_id` INTEGER DEFAULT NULL,
    FOREIGN KEY (`company_id`) REFERENCES `companies` (`id`),
    FOREIGN KEY (`poacher_id`) REFERENCES `companies` (`id`)
);
