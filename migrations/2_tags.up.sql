CREATE TABLE `tags` (
  `id` bigint unsigned PRIMARY KEY,
  `name` text NOT NULL,
  UNIQUE (`name`)
);
