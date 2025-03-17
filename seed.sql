CREATE ROLE anonymous;
CREATE ROLE authenticated;

GRANT anonymous TO authenticated WITH INHERIT TRUE;