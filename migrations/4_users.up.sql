CREATE TABLE `users` (
  `id` bigint unsigned PRIMARY KEY NOT NULL,
  `username` text NOT NULL,
  `password` text NOT NULL,
  `role` text NOT NULL,
  UNIQUE(`username`)
);
