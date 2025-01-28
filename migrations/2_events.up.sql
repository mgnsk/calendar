CREATE TABLE `events` (
  `id` bigint unsigned NOT NULL,
  `unix_timestamp` bigint unsigned NOT NULL,
  `human_timestamp` text NOT NULL,
  `title` text NOT NULL,
  `content` text NOT NULL,
  PRIMARY KEY (`id`)
);
