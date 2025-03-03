{ pkgs, lib, config, inputs, ... }:

{
  packages = with pkgs; [
    git
    wgo
  ];
  env = {
    PALAKIT_ENVIRONMENT = "development";
    PALAKIT_DB_CONNECTION_STRING = "dbname=palakit";
    PALAKIT_AUTH_DISABLE = "1";
  };
  languages.go.enable = true;
  services.postgres = {
    enable = true;
    initialDatabases = [
      { name = "palakit"; }
    ];
  };
  processes.palakit = {
    exec = "wgo run cmd/palakit/palakit.go start";
    process-compose.depends_on.postgres.condition = "process_ready";
  };
}
