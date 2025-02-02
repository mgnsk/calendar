CREATE TABLE `tags` (
  `id` bigint NOT NULL,
  `name` text NOT NULL,
  `event_count` bigint unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE (`name`)
);
