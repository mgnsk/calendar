CREATE TABLE `tags` (
  `id` bigint NOT NULL,
  `name` text NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE (`name`)
);
