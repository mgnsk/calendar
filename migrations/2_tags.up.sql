CREATE TABLE `tags` (
  `id` bigint PRIMARY KEY,
  `name` text NOT NULL,
  UNIQUE (`name`)
);
