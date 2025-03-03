CREATE TABLE `users` (
  `id` bigint PRIMARY KEY,
  `username` text NOT NULL,
  `password` text NOT NULL,
  `role` text NOT NULL,
  UNIQUE(`username`)
);
