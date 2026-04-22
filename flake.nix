{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs, ... }:
    let
      supportedSystems = nixpkgs.lib.systems.flakeExposed;

      forAllSystems =
        f:
        nixpkgs.lib.genAttrs supportedSystems (
          system:
          let
            pkgs = import nixpkgs {
              inherit system;
              # config.allowUnfree = true;
            };
          in
          f { inherit pkgs system; }
        );
    in
    {
      packages = forAllSystems (
        { pkgs, ... }:
        rec {
          todo = pkgs.buildGoModule {
            pname = "todo";
            version = "0.0.1";
            vendorHash = "sha256-GkW+W3Mwv9RqV9XcxKR8zD4q3p/w9mmTRrRuYRR7E9M=";
            src = ./.;
          };

          default = todo;
        }
      );

      devShells = forAllSystems (
        { pkgs, ... }:
        {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [
              go
              gopls
            ];
            shellHook = ''
              export PS1="(todo) $PS1"
            '';
          };
        }
      );

      nixosModules = {
        default = self.nixosModules.todo;

        todo =
          { pkgs, ... }:
          {
            environment.systemPackages = [
              self.packages.${pkgs.stdenv.hostPlatform.system}.todo
            ];
          };
      };

      homeModules = {
        default = self.homeManagerModules.todo;

        todo =
          { pkgs, ... }:
          {
            home.packages = [
              self.packages.${pkgs.stdenv.hostPlatform.system}.todo
            ];
          };
      };
    };
}
