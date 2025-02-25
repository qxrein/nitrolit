# Nitrolit

## Prerequisites

- Ensure you have the necessary dependencies installed for your operating system.
- For Nix-based systems, ensure you have Nix installed.

## Debian-based Systems

### With Nushell

1. **Install Nushell**: Follow the [official Nushell installation guide](https://www.nushell.sh/book/installation.html).

2. **Set Up Environment**:
   ```sh
   #!/usr/bin/env nu
   def --env nitrolit [] {
       cd ~/projects/nitrolit
       nix develop -c ./nitrolit
   }
   export-env {
       $env.PATH = ($env.PATH | append ([$env.PWD]))
   }
   ```

3. **Run Nitrolit**:
   ```sh
   nitrolit
   ```

### Without Nushell

1. **Install Dependencies**:
   ```sh
   sudo apt update
   sudo apt install -y rustc cargo nodejs lm-sensors binutils lld
   ```

2. **Run Nitrolit**:
   ```sh
   cd ~/projects/nitrolit
   ./nitrolit
   ```

## Fedora-based Systems

### With Nushell

1. **Install Nushell**: Follow the [official Nushell installation guide](https://www.nushell.sh/book/installation.html).

2. **Set Up Environment**:
   ```sh
   #!/usr/bin/env nu
   def --env nitrolit [] {
       cd ~/projects/nitrolit
       nix develop -c ./nitrolit
   }
   export-env {
       $env.PATH = ($env.PATH | append ([$env.PWD]))
   }
   ```

3. **Run Nitrolit**:
   ```sh
   nitrolit
   ```

### Without Nushell

1. **Install Dependencies**:
   ```sh
   sudo dnf install -y rust cargo nodejs lm_sensors binutils lld
   ```

2. **Run Nitrolit**:
   ```sh
   cd ~/projects/nitrolit
   ./nitrolit
   ```

## Nix-based Systems

### With Nushell

1. **Install Nushell**: Follow the [official Nushell installation guide](https://www.nushell.sh/book/installation.html).

2. **Set Up Environment**:
   ```sh
   #!/usr/bin/env nu
   def --env nitrolit [] {
       cd ~/projects/nitrolit
       nix develop -c ./nitrolit
   }
   export-env {
       $env.PATH = ($env.PATH | append ([$env.PWD]))
   }
   ```

3. **Run Nitrolit**:
   ```sh
   nitrolit
   ```

### Without Nushell

1. **Set Up Environment**:
   ```sh
   cd ~/projects/nitrolit
   nix develop
   ```

2. **Run Nitrolit**:
   ```sh
   ./nitrolit
   ```


