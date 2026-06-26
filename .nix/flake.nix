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
          cd "$(git rev-parse --show-toplevel)" || exit 1
          SLINT_LIVE_PREVIEW=1 cargo run -p app --features dev-ui
        '';

        script-test = pkgs.writeShellScriptBin "test-rust" ''
          cd "$(git rev-parse --show-toplevel)" || exit 1
          cargo nextest run
        '';

        script-lint = pkgs.writeShellScriptBin "lint" ''
          cd "$(git rev-parse --show-toplevel)" || exit 1
          cargo clippy
        '';

        script-fmt = pkgs.writeShellScriptBin "fmt" ''
          cd "$(git rev-parse --show-toplevel)" || exit 1
          cargo fmt
        '';

        script-build-go = pkgs.writeShellScriptBin "build-go" ''
          root="$(git rev-parse --show-toplevel)"
          cd "$root/sidecars/FB" || exit 1
          go build -o fb_generator.exe ./cmd/report_generator
          echo "Go Sidecar erfolgreich kompiliert (sidecars/FB/fb_generator.exe)"
          cd "$root/sidecars/Vorpruefung" || exit 1
          go build -o vp_generator.exe ./cmd/vp_generator
          echo "Go Sidecar erfolgreich kompiliert (sidecars/Vorpruefung/vp_generator.exe)"
        '';

        script-run-vorpruefung = pkgs.writeShellScriptBin "run-vorpruefung" ''
          root="$(git rev-parse --show-toplevel)"
          cd "$root/sidecars/Vorpruefung" || exit 1
          go build -o vp_generator.exe ./cmd/vp_generator && ./vp_generator.exe
        '';

        # Erzeugt die Vorlage MIT eingetragenem Budget (Blatt "I. Budget").
        # Nur der Generator-Aufruf trägt das Budget ein – testfill tut das NICHT.
        # Inputs:  testdata/budgets/*.xlsx   (Scanner-Inputs)
        #          testdata/fixtures/*.json  (kanonische BudgetData-JSON)
        # Outputs: tmp/
        # Usage: vorpruefung-budget [budget.json] [output.xlsx]
        script-vorpruefung-budget = pkgs.writeShellScriptBin "vorpruefung-budget" ''
          root="$(git rev-parse --show-toplevel)"
          budget="$(readlink -f "''${1:-$root/testdata/fixtures/budget.example.json}")"
          out="$(readlink -f "''${2:-$root/tmp/vp_output.xlsx}")"
          mkdir -p "$(dirname "$out")"
          go run -C "$root/sidecars/Vorpruefung" ./cmd/vp_generator -budget "$budget" -o "$out"
        '';

        # Voller Durchlauf: erst Vorlage mit Budget erzeugen, dann via testfill alle
        # weiteren Daten (Dashboard, KMW, Finanzberichte, Mittelanforderung) befüllen.
        # Usage: vorpruefung-fill [budget.json] [template.xlsx] [output.xlsx]
        script-vorpruefung-fill = pkgs.writeShellScriptBin "vorpruefung-fill" ''
          root="$(git rev-parse --show-toplevel)"
          budget="$(readlink -f "''${1:-$root/testdata/fixtures/budget.example.json}")"
          template="$(readlink -f "''${2:-$root/tmp/vp_template.xlsx}")"
          out="$(readlink -f "''${3:-$root/tmp/vp_befuellt.xlsx}")"
          mkdir -p "$(dirname "$template")"
          go run -C "$root/sidecars/Vorpruefung" ./cmd/vp_generator -budget "$budget" -o "$template" || exit 1
          go run -C "$root/sidecars/Vorpruefung" ./testfill -in "$template" -budget "$budget" -o "$out"
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
          prek

          # --- Custom Scripts ---
          script-dev
          script-test
          script-lint
          script-fmt
          script-build-go
          script-run-vorpruefung
          script-vorpruefung-budget
          script-vorpruefung-fill

          git-filter-repo
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
            echo '  run-vorpruefung - Baut und führt das Vorpruefung Go-Skript aus'
            echo '  vorpruefung-budget [budget.json] [out.xlsx] - Vorlage NUR mit Budget erzeugen'
            echo '  vorpruefung-fill [budget.json] [tmpl.xlsx] [out.xlsx] - Budget + alle Testdaten (testfill)'
            echo '  prek         - Git pre-commit hooks ausführen'
          '';
        };
      });
}
