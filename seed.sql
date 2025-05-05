CREATE ROLE anonymous;
CREATE ROLE authenticated;

GRANT anonymous TO authenticated WITH INHERIT TRUE;

CREATE TABLE document (id SERIAL PRIMARY KEY, body TEXT);
GRANT ALL ON TABLE document TO anonymous;
GRANT USAGE ON SEQUENCE document_id_seq TO anonymous;