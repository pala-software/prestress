{ pkgs, lib, config, inputs, ... }:

{
  packages = with pkgs; [
    git
    wgo
  ];
  env = {
    PALAKIT_ENVIRONMENT = "development";
    PALAKIT_DB_CONNECTION_STRING = "dbname=palakit";
    PALAKIT_AUTH_INTROSPECTION_URL = "http://localhost:8081/introspect";
    PALAKIT_AUTH_CLIENT_ID = "dev";
    PALAKIT_AUTH_CLIENT_SECRET = "dev";
  };
  languages.go.enable = true;
  services.postgres = {
    enable = true;
    initialDatabases = [
      { name = "palakit"; }
      { name = "palakit_test"; }
    ];
  };
  processes.palakit = {
    exec = "wgo run cmd/palakit/palakit.go start";
    process-compose.depends_on.postgres.condition = "process_ready";
  };
}
