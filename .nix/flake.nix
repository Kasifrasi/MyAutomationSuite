{
  description = "A pure Nix flake for a Rust/Slint GUI and Go Sidecar project";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    rust-overlay.url = "github:oxalica/rust-overlay";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, rust-overlay, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        overlays = [ (import rust-overlay) ];
        pkgs = import nixpkgs {
          inherit system overlays;
        };

        # Rust toolchain setup
        rustToolchain = pkgs.rust-bin.stable.latest.default.override {
          extensions = [ "rust-src" "rust-analyzer" "clippy" ];
        };

        # Libraries required at runtime by wayland/opengl/etc. (Slint GUI)
        runtimeLibs = with pkgs; [
          wayland
          libxkbcommon
          libGL
          fontconfig
          libX11
          libXcursor
          libXi
          libXrandr
        ];

        # --- Echte Helper-Skripte statt Bash-Aliase ---
        # Diese landen direkt im $PATH der Nix-Shell und funktionieren in JEDER Shell (auch Nushell).
        script-dev = pkgs.writeShellScriptBin "dev" ''
          SLINT_LIVE_PREVIEW=1 cargo run -p app --features dev-ui
        '';

        script-test = pkgs.writeShellScriptBin "test-rust" ''
          cargo nextest run
        '';

        script-lint = pkgs.writeShellScriptBin "lint" ''
          cargo clippy
        '';

        script-fmt = pkgs.writeShellScriptBin "fmt" ''
          cargo fmt
        '';

        script-build-go = pkgs.writeShellScriptBin "build-go" ''
          cd sidecars/Excelize && go build -o scanner.exe ./cmd/generator
          echo "Go Sidecar erfolgreich kompiliert (sidecars/Excelize/scanner.exe)"
        '';

        # Build-time tools and dependencies
        buildInputs = with pkgs; [
          # --- C/C++ Build Tools ---
          pkg-config
          mold
          clang

          # --- Go Dependencies (Sidecar) ---
          go
          gopls       # Go language server
          gotools     # Go tools like goimports
          go-tools    # staticcheck

          # --- Rust Tools (GUI & Core Scanner) ---
          bacon
          cargo-release
          cargo-about
          cargo-audit
          cargo-cyclonedx
          cargo-deny
          cargo-edit
          cargo-expand
          cargo-license
          cargo-llvm-cov
          cargo-nextest
          sccache
          slint-lsp

          # --- Custom Scripts ---
          script-dev
          script-test
          script-lint
          script-fmt
          script-build-go
        ] ++ runtimeLibs;

      in {
        devShells.default = pkgs.mkShell {
          buildInputs = buildInputs;

          nativeBuildInputs = [ rustToolchain ];

          # Environment variables
          CC = "clang";
          CXX = "clang++";
          CARGO_TARGET_X86_64_UNKNOWN_LINUX_GNU_LINKER = "clang";
          CARGO_TARGET_X86_64_UNKNOWN_LINUX_GNU_RUSTFLAGS = "-C link-arg=-fuse-ld=mold";
          RUSTC_WRAPPER = "sccache";

          # Ensure runtime libraries can be found by dynamically loaded libraries
          LD_LIBRARY_PATH = "${pkgs.lib.makeLibraryPath runtimeLibs}:/run/opengl-driver/lib:/run/opengl-driver-32/lib";

          # Shell hook to run when entering the shell
          shellHook = ''
            echo '==================================================='
            echo '🦀 Rust & Go Workspace | Automation Suite'
            echo '==================================================='

            go version
            cargo --version

            echo ""
            echo 'Available commands:'
            echo '  dev          - SLINT_LIVE_PREVIEW=1 cargo run -p app --features dev-ui'
            echo '  test-rust    - cargo nextest run'
            echo '  lint         - cargo clippy'
            echo '  build-go     - Kompiliert das Go Sidecar'
          '';
        };
      });
}
