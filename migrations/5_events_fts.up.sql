-- And an external content fts5 table.
CREATE VIRTUAL TABLE events_fts_idx USING fts5(title, description, content='events', content_rowid='id', tokenize='trigram case_sensitive 0 remove_diacritics 1');

-- Triggers to keep the FTS index up to date.
CREATE TRIGGER events_ai AFTER INSERT ON events BEGIN
  INSERT INTO events_fts_idx(rowid, title, description) VALUES (new.id, new.title, new.description);
END;
CREATE TRIGGER events_ad AFTER DELETE ON events BEGIN
  INSERT INTO events_fts_idx(events_fts_idx, rowid, title, description) VALUES('delete', old.id, old.title, old.description);
END;
CREATE TRIGGER events_au AFTER UPDATE ON events BEGIN
  INSERT INTO events_fts_idx(events_fts_idx, rowid, title, description) VALUES('delete', old.id, old.title, old.description);
  INSERT INTO events_fts_idx(rowid, title, description) VALUES (new.id, new.title, new.description);
END;
