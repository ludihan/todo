# todo

todo is a tool for quick reminders and taking notes.

## Usage

TODO

## Installation
### Linux, Windows or Mac:
1. Install [Go](https://go.dev/)
2. Run:
```
go install github.com/ludihan/todo@latest
```
### Nix or NixOS (using flakes):
1. Enable flakes
2. Add this repo to the input of your flake:
```nix
inputs = {
  nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  todo = {
    url = "github:ludihan/todo";
    inputs.nixpkgs.follows = "nixpkgs";
  };
};
```
3. Add the program as a module to your NixOS or Home Manager config:
#### NixOS:
```nix
outputs = { self, nixpkgs, todo, ... }:
{
  nixosConfigurations.example = nixpkgs.lib.nixosSystem {
    inherit system;
    modules = [
      todo.nixosModules.todo
    ];
  };
}
```

#### Home Manager:
```nix
outputs = { self, nixpkgs, todo, ... }:
{
  homeConfigurations.example = home-manager.lib.homeManagerConfiguration {
    inherit system;
    modules = [
      todo.homeModules.todo
    ];
  };
}
```
