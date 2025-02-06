CREATE TABLE `users` (
  `id` bigint PRIMARY KEY NOT NULL,
  `username` text NOT NULL,
  `password` text NOT NULL,
  `role` tinyint NOT NULL,
  UNIQUE(`username`)
);
