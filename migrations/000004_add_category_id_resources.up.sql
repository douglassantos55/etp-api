ALTER TABLE `resources`
ADD COLUMN `category_id` INTEGER
REFERENCES `categories`(`id`)
