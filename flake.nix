{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    {
      nixpkgs,
      ...
    }:
    let
      supportedSystems = nixpkgs.lib.systems.flakeExposed;
      forAllSystems =
        function:
        nixpkgs.lib.genAttrs supportedSystems (
          system:
          function (
            import nixpkgs {
              inherit system;
              config.allowUnfree = true;
            }
          )
        );
    in
    {
      packages = forAllSystems (
        { pkgs, ... }:
        rec {
          default = todo;
          todo = pkgs.buildGoModule {
            pname = "todo";
            version = "0.0.1";
            vendorHash = "sha256-8uKnWsQZEzLV5t8wtYqAONrsIRWdZx7OF7DAeOMckTU=";
            src = ./.;
          };
        }
      );

      devShells = forAllSystems (
        { pkgs, ... }:
        {
          default =
            with pkgs;
            mkShell {
              buildInputs = [
                go
                gopls
              ];

              shellHook = ''
                export PS1="(todo) $PS1"
              '';
            };
        }
      );
    };
}
