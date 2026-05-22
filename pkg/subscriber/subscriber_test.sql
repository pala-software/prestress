DROP SCHEMA IF EXISTS test CASCADE;
CREATE SCHEMA test;
GRANT USAGE ON SCHEMA test TO public;

CREATE TABLE test.document (body TEXT);
INSERT INTO test.document (body) VALUES (NULL), ('1'), ('2');
GRANT SELECT ON TABLE test.document TO public;
