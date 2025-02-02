CREATE VIRTUAL TABLE events_fts USING fts5(title, description, url, tags, id UNINDEXED, tokenize="trigram case_sensitive 0 remove_diacritics 1");
