DROP SCHEMA IF EXISTS crud CASCADE;
CREATE SCHEMA crud;
GRANT USAGE ON SCHEMA crud TO public;

CREATE TABLE crud.find (value TEXT);
INSERT INTO crud.find (value) VALUES (NULL), ('1'), ('2');
GRANT SELECT ON TABLE crud.find TO public;

CREATE TABLE crud.create (value TEXT);
INSERT INTO crud.create (value) VALUES (NULL), ('1'), ('2');
GRANT SELECT ON TABLE crud.create TO public;
GRANT INSERT ON TABLE crud.create TO public;

CREATE TABLE crud.update (value TEXT);
INSERT INTO crud.update (value) VALUES (NULL), ('1'), ('2');
GRANT SELECT ON TABLE crud.update TO public;
GRANT UPDATE ON TABLE crud.update TO public;

CREATE TABLE crud.delete (value TEXT);
INSERT INTO crud.delete (value) VALUES (NULL), ('1'), ('2');
GRANT SELECT ON TABLE crud.delete TO public;
GRANT DELETE ON TABLE crud.delete TO public;
