{ pkgs, lib, config, inputs, ... }:

{
  packages = with pkgs; [
    git
    wgo
    httpie
  ];
  env = {
    PRESTRESS_DB = "dbname=prestress_dev";
    PRESTRESS_TEST_DB = "dbname=prestress_test";
    PRESTRESS_ALLOWED_ORIGINS = "*";
    PRESTRESS_OAUTH_INTROSPECTION_URL = "http://localhost:8081/introspect";
    PRESTRESS_OAUTH_CLIENT_ID = "dev";
    PRESTRESS_OAUTH_CLIENT_SECRET = "dev";
    OTEL_SERVICE_NAME = "prestress";
    OTEL_EXPORTER_OTLP_ENDPOINT = "http://localhost:4318";
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
  devcontainer = {
    enable = true;
    settings.customizations.vscode.extensions = [
      "mkhl.direnv"
      "streetsidesoftware.code-spell-checker"
      "eamodio.gitlens"
      "golang.go"
      "bbenoist.Nix"
      "redhat.vscode-yaml"
    ];
  };
}
