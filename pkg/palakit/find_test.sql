DROP SCHEMA IF EXISTS find_test CASCADE;
CREATE SCHEMA find_test;

CREATE TABLE find_test.test (test TEXT);

INSERT INTO find_test.test (test) VALUES ('1'), ('2');

GRANT USAGE ON SCHEMA find_test TO public;
GRANT SELECT ON TABLE find_test.test TO public;