CREATE TABLE `events_tags` (
  `tag_id` bigint NOT NULL,
  `event_id` bigint NOT NULL,
  PRIMARY KEY (`tag_id`, `event_id`)
);
CREATE INDEX events_tags_event_id_idx ON events_tags (event_id);
