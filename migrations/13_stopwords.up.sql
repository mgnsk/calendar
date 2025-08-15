CREATE TABLE `stopwords` (
  `id` bigint PRIMARY KEY,
  `word` text NOT NULL,
  UNIQUE (`word`)
);
