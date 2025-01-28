CREATE TABLE `tags` (
  `id` bigint unsigned NOT NULL,
  `name` text NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE (`name`)
);
