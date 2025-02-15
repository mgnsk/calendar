CREATE TABLE `events` (
  `id` bigint unsigned PRIMARY KEY,
  `start_at_unix` bigint NOT NULL,
  `end_at_unix` bigint,
  `tz_offset` int NOT NULL,
  `title` text NOT NULL,
  `description` text NOT NULL,
  `url` text NOT NULL
);

CREATE INDEX events_timestamp_idx ON events (start_at_unix);
