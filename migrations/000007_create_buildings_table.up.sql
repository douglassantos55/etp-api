CREATE TABLE IF NOT EXISTS `buildings` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `name` VARCHAR(255) NOT NULL,
    `downtime` INTEGER DEFAULT NULL,
    `wages_per_hour` INTEGER UNSIGNED DEFAULT 0,
    `admin_per_hour` INTEGER UNSIGNED DEFAULT 0,
    `maintenance_per_hour` INTEGER UNSIGNED DEFAULT 0,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS `buildings_requirements` (
    `building_id` INTEGER NOT NULL,
    `resource_id` INTEGER NOT NULL,
    `qty` INTEGER UNSIGNED NOT NULL,
    `quality` TINYINT UNSIGNED NOT NULL,
    PRIMARY KEY (`building_id`, `resource_id`),
    FOREIGN KEY (`building_id`) REFERENCES `buildings`(`id`),
    FOREIGN KEY (`resource_id`) REFERENCES `resources`(`id`)
);

CREATE TABLE IF NOT EXISTS `buildings_resources` (
    `building_id` INTEGER NOT NULL,
    `resource_id` INTEGER NOT NULL,
    `qty_per_hour` INTEGER UNSIGNED NOT NULL,
    PRIMARY KEY (`building_id`, `resource_id`),
    FOREIGN KEY (`building_id`) REFERENCES `buildings`(`id`),
    FOREIGN KEY (`resource_id`) REFERENCES `resources`(`id`)
);

CREATE TABLE IF NOT EXISTS `companies_buildings` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `name` VARCHAR(255) NOT NULL,
    `company_id` INTEGER NOT NULL,
    `building_id` INTEGER NOT NULL,
    `level` TINYINT UNSIGNED DEFAULT 1,
    `position` TINYINT UNSIGNED DEFAULT NULL,
    `built_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `demolished_at` TIMESTAMP DEFAULT NULL,
    FOREIGN KEY (`company_id`) REFERENCES `companies`(`id`),
    FOREIGN KEY (`building_id`) REFERENCES `buildings`(`id`)
);
