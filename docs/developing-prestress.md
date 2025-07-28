# Developing Prestress

Follow the following steps to get started and familiar with developing
Prestress. This is not a guide for using Prestress, but instead how to develop
Prestress itself.

1. Devcontainer is supported by this repository, so use that with Visual Studio
   Code for quick setup of development environment.

2. To start Prestress, run the following command:

   ```sh
   devenv up
   ```

3. You can access development database with following command:

   ```
   psql prestress_dev
   ```

4. Database is migrated automatically at startup. To start fresh, you can stop
   processes and remove the whole PostgreSQL state by deleting directory at
   `.devenv/state/postgres`:

   ```sh
   rm .devenv/state/postgres
   ```

5. To recreate database structure you can use `psql prestress_dev` command. But
   remember to write automated tests too when making making changes to
   Prestress.

   The following example schema applied by running `psql prestress_dev` and
   pasting it to the console:

   ```sql
   CREATE TABLE document (id SERIAL PRIMARY KEY, body TEXT);
   GRANT ALL ON TABLE document TO anonymous;
   GRANT USAGE ON SEQUENCE document_id_seq TO anonymous;
   ```

6. To send HTTP requests from terminal, you can use httpie which is preinstalled
   with the development environment:

   ```sh
   # Create row on document table:
   http POST http://localhost:8080/public/document body="Hello, World!"

   # Retrieve 10 rows from document table:
   http GET http://localhost:8080/public/document?limit=10
   ```

7. To test out authentication, you can use mock identity provider which comes
   with the project. Just start it and you will be prompted for wanted role and
   user ID (token subject). After inputting those you'll get access token that
   you can use for HTTP requests. You have to keep the command running because
   it acts as a server.

   ```sh
   go run cmd/idp/idp.go
   Enter wanted role name: authenticated
   Enter wanted user ID: user-1
   Created token:
   TOKEN-1

   Enter wanted role name:
   ```

8. Again, you can use httpie to send authenticated requests by adding
   authorization header with the generated token:

   ```
   http GET http://localhost:8080/public/document Authorization:"Bearer TOKEN-1"
   ```
