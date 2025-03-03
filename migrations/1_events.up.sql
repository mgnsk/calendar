CREATE TABLE `events` (
  `id` bigint unsigned PRIMARY KEY,
  `start_at_unix` bigint NOT NULL,
  `tz_offset` int NOT NULL,
  `title` text NOT NULL,
  `description` text NOT NULL,
  `url` text NOT NULL,
  `location` text NOT NULL,
  `latitude` real NOT NULL,
  `longitude` real NOT NULL,
  `is_draft` tinyint NOT NULL,
  `user_id` bigint unsigned NOT NULL
);
