CREATE TABLE IF NOT EXISTS `resources` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `name` VARCHAR(255) NOT NULL,
    `image` VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS `resources_requirements` (
    `resource_id` INTEGER NOT NULL,
    `requirement_id` INTEGER NOT NULL,
    `qty` INTEGER UNSIGNED NOT NULL,
    `quality` TINYINT UNSIGNED NOT NULL,
    PRIMARY KEY (`resource_id`, `requirement_id`),
    FOREIGN KEY (`resource_id`) REFERENCES `resources`(`id`),
    FOREIGN KEY (`requirement_id`) REFERENCES `resources`(`id`)
);
