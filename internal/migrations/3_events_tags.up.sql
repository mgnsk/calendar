CREATE TABLE `events_tags` (
  `tag_id` bigint NOT NULL,
  `event_id` bigint NOT NULL,
  PRIMARY KEY (`event_id`, `tag_id`)
);
