{ pkgs ? import <nixpkgs> {} }:

pkgs.stdenv.mkDerivation rec {
  pname = "nitroctl";
  version = "1.0.0";

  src = ./.;

  installPhase = ''
    mkdir -p $out/bin
    cp ${./nitroctl} $out/bin/nitroctl
    chmod +x $out/bin/nitroctl
  '';
}

