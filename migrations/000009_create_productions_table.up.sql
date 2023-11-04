CREATE TABLE IF NOT EXISTS `productions` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `resource_id` INTEGER NOT NULL,
    `building_id` INTEGER NOT NULL,
    `qty` INTEGER UNSIGNED NOT NULL,
    `quality` TINYINT UNSIGNED NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `finishes_at` TIMESTAMP,
    FOREIGN KEY (`resource_id`) REFERENCES `resources` (`id`),
    FOREIGN KEY (`building_id`) REFERENCES `companies_buildings` (`id`)
);
