CREATE TABLE IF NOT EXISTS `researches` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `patents` TINYINT DEFAULT 0,
    `investment` INTEGER DEFAULT 0,
    `finishes_at` TIMESTAMP,
    `completed_at` TIMESTAMP,
    `company_id` INTEGER NOT NULL,
    `resource_id` INTEGER NOT NULL,
    FOREIGN KEY (`resource_id`) REFERENCES `resources` (`id`),
    FOREIGN KEY (`company_id`) REFERENCES `companies` (`id`)
);

CREATE TABLE IF NOT EXISTS `assigned_staff` (
    `staff_id` INTEGER NOT NULL,
    `research_id` INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS `resources_qualities` (
    `resource_id` INTEGER NOT NULL,
    `company_id` INTEGER NOT NULL,
    `quality` TINYINT UNSIGNED NOT NULL,
    `patents` TINYINT DEFAULT 0,
    PRIMARY KEY (`resource_id`, `company_id`),
    FOREIGN KEY (`resource_id`) REFERENCES `resources`(`id`),
    FOREIGN KEY (`company_id`) REFERENCES `companies`(`id`)
);

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

CREATE TABLE IF NOT EXISTS `staff_searches` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `company_id` INTEGER NOT NULL,
    `finishes_at` TIMESTAMP,
    `started_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`company_id`) REFERENCES `companies` (`id`)
);

CREATE TABLE IF NOT EXISTS `trainings` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `result` TINYINT,
    `staff_id` INTEGER,
    `company_id` INTEGER,
    `investment` INTEGER,
    `finishes_at` TIMESTAMP,
    `completed_at` TIMESTAMP,
    FOREIGN KEY (`staff_id`) REFERENCES `research_staff` (`id`),
    FOREIGN KEY (`company_id`) REFERENCES `companies` (`id`)
);
