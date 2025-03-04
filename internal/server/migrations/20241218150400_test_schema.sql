CREATE SCHEMA test;

CREATE TABLE test.document (
  id SERIAL PRIMARY KEY,
  body TEXT,
  owner TEXT,
  public BOOLEAN
);

ALTER TABLE test.document ENABLE ROW LEVEL SECURITY; 

CREATE POLICY "Everyone can see public documents"
ON test.document
FOR SELECT
USING (public = TRUE);

CREATE POLICY "Owner can see their own documents"
ON test.document
FOR SELECT
USING (owner = auth.uid());

CREATE VIEW test.owned_document
WITH ( security_invoker = TRUE )
AS SELECT * FROM test.document WHERE owner = auth.uid();

CREATE VIEW test.private_document
WITH ( security_invoker = TRUE )
AS SELECT * FROM test.document WHERE public = FALSE;

INSERT INTO test.document (body, owner, public)
VALUES
  ('Hello, World!', 'user-1', TRUE),
  ('Curse you, world!', 'user-1', FALSE),
  ('My profile', 'user-2', TRUE),
  ('My private things', 'user-2', FALSE);

CREATE ROLE anonymous;
GRANT USAGE ON SCHEMA test TO anonymous;
GRANT SELECT ON TABLE test.document TO anonymous;

CREATE ROLE authenticated;
GRANT USAGE ON SCHEMA test TO authenticated;
GRANT SELECT ON TABLE test.document TO authenticated;
GRANT SELECT ON TABLE test.owned_document TO authenticated;
GRANT SELECT ON TABLE test.private_document TO authenticated;