CREATE TABLE `events` (
  `id` bigint PRIMARY KEY NOT NULL,
  `start_at_unix` bigint NOT NULL,
  `start_at_rfc3339` text NOT NULL,
  `end_at_unix` bigint,
  `end_at_rfc3339` text,
  `title` text NOT NULL,
  `description` text NOT NULL,
  `url` text NOT NULL,
  `fts_data` text NOT NULL
);

CREATE INDEX events_timestamp_idx ON events (start_at_unix);
