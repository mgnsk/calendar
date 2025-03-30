CREATE TABLE `invites` (
  `token` text PRIMARY KEY,
  `valid_until_unix` bigint NOT NULL,
  `created_by` bigint NOT NULL
);
