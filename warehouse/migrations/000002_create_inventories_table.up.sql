CREATE TABLE IF NOT EXISTS `inventories` (
    `resource_id` INTEGER,
    `company_id` INTEGER,
    `quantity` INTEGER UNSIGNED,
    `quality` TINYINT UNSIGNED,
    `sourcing_cost` DECIMAL(10,2),
    PRIMARY KEY (`resource_id`, `company_id`, `quality`)
)
