CREATE TABLE `events_tags` (
  `tag_id` integer NOT NULL,
  `event_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`tag_id`, `event_id`)
);
