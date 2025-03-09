{ pkgs, lib, config, inputs, ... }:

{
  packages = with pkgs; [
    git
    wgo
  ];
  env = {
    PRESTRESS_ENVIRONMENT = "development";
    PRESTRESS_DB_CONNECTION_STRING = "dbname=prestress_dev";
    PRESTRESS_AUTH_INTROSPECTION_URL = "http://localhost:8081/introspect";
    PRESTRESS_AUTH_CLIENT_ID = "dev";
    PRESTRESS_AUTH_CLIENT_SECRET = "dev";
  };
  languages.go.enable = true;
  services.postgres = {
    enable = true;
    initialDatabases = [
      { name = "prestress_dev"; schema = ./seed.sql; }
      { name = "prestress_test"; schema = ./seed.sql; }
    ];
  };
  processes.prestress = {
    exec = ''
      go run cmd/prestress/prestress.go migrate && \
      wgo run cmd/prestress/prestress.go start
    '';
    process-compose.depends_on.postgres.condition = "process_healthy";
  };
}
