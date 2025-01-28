CREATE TABLE `events` (
  `id` bigint NOT NULL,
  `unix_timestamp` bigint NOT NULL,
  `rfc3339_timestamp` text NOT NULL,
  `title` text NOT NULL,
  `content` text NOT NULL,
  PRIMARY KEY (`id`)
);
