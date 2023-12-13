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
