#!/usr/bin/env nu

def --env nitrolit [] {
    cd ~/projects/nitrolit
    nix develop -c ./nitrolit
}
export-env {
    $env.PATH = ($env.PATH | append ([$env.PWD]))
}

