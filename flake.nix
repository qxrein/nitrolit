{
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs = { self, nixpkgs }:
  let
    system = "x86_64-linux";
    pkgs = import nixpkgs { inherit system; };
  in {
    devShells.${system}.default = pkgs.mkShell {
      buildInputs = [
        pkgs.rustc
        pkgs.cargo
        pkgs.cargo-tauri
        pkgs.lm_sensors
        pkgs.nodejs
        pkgs.binutils
        pkgs.lld
      ];

      shellHook = ''
        export PATH=${pkgs.coreutils}/bin:${pkgs.rustc}/bin:${pkgs.cargo}/bin:${pkgs.cargo-tauri}/bin:${pkgs.lm_sensors}/bin:${pkgs.nodejs}/bin:${pkgs.binutils}/bin:${pkgs.lld}/bin:$PATH
      '';
    };
  };
}
