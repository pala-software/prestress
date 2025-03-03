{ pkgs, lib, config, inputs, ... }:

{
  packages = [ pkgs.git ];
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
}
