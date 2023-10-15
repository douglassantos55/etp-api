CREATE TABLE IF NOT EXISTS `inventories` (
    `resource_id` INTEGER,
    `company_id` INTEGER,
    `quantity` INTEGER UNSIGNED,
    `quality` TINYINT UNSIGNED,
    `sourcing_cost` INTEGER UNSIGNED,
    PRIMARY KEY (`resource_id`, `company_id`, `quality`),
    FOREIGN KEY (`resource_id`) REFERENCES `resources`(`id`)
)
