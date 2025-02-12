CREATE TABLE `events_tags` (
  `tag_id` integer NOT NULL,
  `event_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`event_id`, `tag_id`)
);
