DROP SCHEMA IF EXISTS test CASCADE;
CREATE SCHEMA test;
GRANT USAGE ON SCHEMA test TO public;

CREATE TABLE test.find (value TEXT);
INSERT INTO test.find (value) VALUES ('1'), ('2');
GRANT SELECT ON TABLE test.find TO public;

CREATE TABLE test.create (value TEXT);
INSERT INTO test.create (value) VALUES ('1'), ('2');
GRANT SELECT ON TABLE test.create TO public;
GRANT INSERT ON TABLE test.create TO public;

CREATE TABLE test.update (value TEXT);
INSERT INTO test.update (value) VALUES ('1'), ('2');
GRANT SELECT ON TABLE test.update TO public;
GRANT UPDATE ON TABLE test.update TO public;

CREATE TABLE test.delete (value TEXT);
INSERT INTO test.delete (value) VALUES ('1'), ('2');
GRANT SELECT ON TABLE test.delete TO public;
GRANT DELETE ON TABLE test.delete TO public;