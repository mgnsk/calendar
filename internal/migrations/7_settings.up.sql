CREATE TABLE `settings` (
  `id` bigint PRIMARY KEY NOT NULL,
  `is_initialized` bool NOT NULL,
  `title` text NOT NULL,
  `description` text NOT NULL,
  `base_url` text NOT NULL,
  `session_secret` blob NOT NULL
);
