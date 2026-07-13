DROP SCHEMA IF EXISTS subscriber CASCADE;
CREATE SCHEMA subscriber;
GRANT USAGE ON SCHEMA subscriber TO public;

CREATE TABLE subscriber.document (body TEXT);
INSERT INTO subscriber.document (body) VALUES (NULL), ('1'), ('2');
GRANT SELECT ON TABLE subscriber.document TO public;
