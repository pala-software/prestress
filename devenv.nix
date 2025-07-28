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
    PRESTRESS_OTEL_TRACES_ENABLE = "0";
    PRESTRESS_OTEL_METRICS_ENABLE = "0";
    PRESTRESS_OTEL_LOGS_ENABLE = "0";
    OTEL_SERVICE_NAME = "prestress";
    OTEL_EXPORTER_OTLP_TRACES_ENDPOINT = "http://localhost:4318/v1/traces";
    OTEL_EXPORTER_OTLP_METRICS_ENDPOINT = "http://localhost:9009/otlp/v1/metrics";
    OTEL_EXPORTER_OTLP_LOGS_ENDPOINT = "http://localhost:3100/otlp/v1/logs";
  };
  languages.go.enable = true;
  services.postgres = {
    enable = true;
    initialScript = builtins.readFile ./init.sql;
    initialDatabases = [
      { name = "prestress_dev"; }
      { name = "prestress_test"; }
    ];
  };
  processes.prestress = {
    exec = ''
      go run ./cmd/prestress migrate && \
      wgo run ./cmd/prestress start
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
