-- And an external content fts5 table.
CREATE VIRTUAL TABLE events_fts USING fts5(title, description, location, content='events', content_rowid='id', tokenize='trigram case_sensitive 0 remove_diacritics 1');

-- Triggers to keep the FTS index up to date.
CREATE TRIGGER events_ai AFTER INSERT ON events BEGIN
  INSERT INTO events_fts(rowid, title, description, location) VALUES (new.id, new.title, new.description, new.location);
END;
CREATE TRIGGER events_ad AFTER DELETE ON events BEGIN
  INSERT INTO events_fts(events_fts, rowid, title, description, location) VALUES('delete', old.id, old.title, old.description, old.location);
END;
CREATE TRIGGER events_au AFTER UPDATE ON events BEGIN
  INSERT INTO events_fts(events_fts, rowid, title, description, location) VALUES('delete', old.id, old.title, old.description, old.location);
  INSERT INTO events_fts(rowid, title, description, location) VALUES (new.id, new.title, new.description, new.location);
END;
