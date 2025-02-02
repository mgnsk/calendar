CREATE TABLE `events` (
  `id` bigint NOT NULL,
  `start_at_unix` bigint NOT NULL,
  `start_at_rfc3339` text NOT NULL,
  `end_at_unix` bigint,
  `end_at_rfc3339` text,
  `title` text NOT NULL,
  `description` text NOT NULL,
  `url` text NOT NULL,
  `fts_data` text NOT NULL,
  PRIMARY KEY (`id`)
);
