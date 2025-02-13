CREATE TABLE `tags` (
  `id` bigint unsigned PRIMARY KEY,
  `name` text NOT NULL,
  `event_count` bigint unsigned NOT NULL,
  UNIQUE (`name`)
);

CREATE INDEX tags_event_count_idx ON tags (event_count);
