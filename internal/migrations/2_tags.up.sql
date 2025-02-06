CREATE TABLE `tags` (
  `id` bigint PRIMARY KEY NOT NULL,
  `name` text NOT NULL,
  `event_count` bigint unsigned NOT NULL,
  UNIQUE (`name`)
);
