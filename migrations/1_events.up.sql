CREATE TABLE `events` (
  `id` bigint unsigned NOT NULL,
  `unix_timestamp` bigint NOT NULL,
  `human_timestamp` text NOT NULL,
  `title` text NOT NULL,
  `content` text NOT NULL,
  PRIMARY KEY (`id`),
  INDEX (`unix_timestamp`)
);
