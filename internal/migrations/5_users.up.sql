CREATE TABLE `users` (
  `id` bigint NOT NULL,
  `username` text NOT NULL,
  `password` text NOT NULL,
  `role` tinyint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE(`username`)
);
