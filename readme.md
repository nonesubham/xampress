# XamPress

<p align="center"><img src="https://socialify.git.ci/nonesubham/xampress/image?font=Source%20Code%20Pro&amp;language=1&amp;name=1&amp;owner=1&amp;pattern=Plus&amp;stargazers=1&amp;theme=Auto" alt="project-image"></p>

**XamPress** is a standalone Command Line Interface (CLI) tool designed to simplify creating, running, cloning, and managing local **WordPress** and **Laravel** projects.

Unlike traditional setups that require installing and configuring XAMPP manually, XamPress supports **two modes** chosen on first run:
1. **Standalone Mode (Recommended)**: Automatically downloads and configures portable **PHP**, **MySQL/MariaDB**, **WP-CLI**, and **Composer** with a live progress bar into `~/.xampress`, stores projects in `~/.xampress/sites`, and serves them directly using PHP's built-in web server (`php -S`).
2. **XAMPP Mode**: Integrates with an existing local [XAMPP](https://www.apachefriends.org/) installation.

---

## Features

- **WordPress & Laravel Support**: Create either a **WordPress** site or a **Laravel** application interactively with `xampress create <name>`.
- **Wrapped Laravel Artisan CLI (`xampress artisan`)**: Run `xampress artisan <command>` using XamPress's portable PHP runtime—automatically validates that the project directory contains a valid `xampress.json` (`"type": "laravel"`) and `artisan` script before executing.
- **Zero-Dependency Standalone Mode**: Works out of the box from any directory—automatically downloads portable PHP 8.3, MySQL/MariaDB, WP-CLI, and Composer into `~/.xampress` with a progress bar.
- **Interactive First-Run Wizard**: Prompts you on first launch to choose between **Standalone Mode** or **XAMPP Mode**.
- **Built-in PHP Development Server**: Start any WordPress or Laravel project on a configurable port (`xampress server start <site> --port 8000`) or expose it on your local network (`0.0.0.0`) with `-n`.
- **Automated Project Creation, Cloning & Deletion**: Set up, clone, or remove complete WordPress/Laravel installations and MySQL databases with a single command.

---

## Release Files

Download the prebuilt binary for your operating system from **[Releases](https://github.com/nonesubham/xampress/releases/)**.

---

## Build From Source

### Prerequisites
- [Go 1.22+](https://go.dev/doc/install)

### Steps
```bash
# Clone the repository
git clone https://github.com/nonesubham/xampress.git
cd xampress

# Install dependencies
go mod tidy

# Build the binary
go build -o xampress main.go
```

Move the `xampress` (or `xampress.exe` on Windows) binary anywhere in your system `PATH` (for example `/usr/local/bin` on Linux/macOS or a directory in your Windows `PATH`).

---

## First-Run Setup

The first time you run any `xampress` command, the interactive setup wizard will launch:

```text
========================================================
           Welcome to XamPress First-Run Setup          
========================================================

Choose how you want to run XamPress:
  [1] Standalone Mode (Recommended)
      Downloads portable PHP, MySQL & WP-CLI into ~/.xampress
      Uses built-in PHP server (`php -S`) and stores sites in ~/.xampress/sites
  [2] XAMPP Mode
      Uses an existing local XAMPP installation

Select mode [1/2] (default: 1):
```

You can re-run this setup wizard at any time with:
```bash
xampress config --setup
```

---

## Usage & Examples

### 1. Create a New WordPress or Laravel Project
```bash
xampress create my-project
```
XamPress will prompt you to choose the project framework:
```text
Select project type:
  1. WordPress
  2. Laravel
Enter choice [1/2] (default: 1):
```
- **WordPress (`1`)**: Downloads WordPress core via WP-CLI, creates the MySQL database, configures `wp-config.php`, installs WordPress, and writes `xampress.json`.
- **Laravel (`2`)**: Automatically downloads `composer.phar` (if not yet installed), scaffolds `laravel/laravel`, creates the MySQL database, configures `.env`, writes `xampress.json` (`"type": "laravel"`), and runs initial migrations (`php artisan migrate`).

### 2. Run Laravel Artisan Commands (`xampress artisan`)
XamPress wraps Laravel's `artisan` CLI and **only executes if the target project directory is validated as a Laravel project via its `xampress.json` file**:
```bash
# Run from inside the Laravel project directory
cd ~/.xampress/sites/my-laravel-app
xampress artisan migrate
xampress artisan make:model Post -mcr

# Or run from anywhere by passing --site / -s
xampress artisan --site my-laravel-app route:list
xampress art -s my-laravel-app make:controller PostController
```

### 3. Start / Stop the Server for a Project
Start the built-in PHP server (`php -S`) and MySQL database for any WordPress or Laravel project:
```bash
# Start on default port (http://localhost:8000)
xampress server start my-project

# Start on a custom port (http://localhost:8080)
xampress server start my-project --port 8080

# Expose server on local network (0.0.0.0) using -n / --network
xampress server start my-project -n

# Run in foreground (Ctrl+C to stop)
xampress server start my-project --foreground
```

Stop the running PHP server and database:
```bash
xampress server stop
```

### 4. List All Installed Projects
```bash
xampress list
```

### 5. Clone an Existing Project
```bash
xampress clone my-project my-project-staging
```

### 6. Delete a Project
```bash
xampress delete my-project
```

### 7. View or Update Configuration
```bash
# View current configuration
xampress config

# Update credentials or default port
xampress config --wp-user admin --wp-pass secret123 --wp-email dev@example.com --port 8080
```

#### Available `config` Flags:
| Flag | Description |
| :--- | :--- |
| `--setup` | Re-run the interactive first-time setup wizard |
| `--mode` | Switch mode (`standalone` or `xampp`) |
| `--port` | Set default PHP server port (default: `8000`) |
| `--sites-dir` | Set custom directory for projects |
| `--xampp-dir` | Set XAMPP installation directory (when in `xampp` mode) |
| `--sql-user` | Set MySQL username (default: `root`) |
| `--sql-pass` | Set MySQL password |
| `--sql-host` | Set MySQL host (default: `127.0.0.1`) |
| `--sql-port` | Set MySQL port (default: `3307` in standalone, `3306` in XAMPP) |
| `--wp-user` | Set default WordPress admin username |
| `--wp-pass` | Set default WordPress admin password |
| `--wp-email` | Set default WordPress admin email |

---

## Project Validation File (`xampress.json`)

Every project created by XamPress contains a `xampress.json` metadata file in its root directory:
```json
{
  "name": "my-project",
  "type": "laravel",
  "db_name": "my_project",
  "created_at": "2026-10-20T22:05:00+05:30"
}
```
When running `xampress artisan`, XamPress validates that `xampress.json` exists in the project directory with `"type": "laravel"` alongside the `artisan` script before executing any command.

---

## Directory Structure (`~/.xampress`)

In **Standalone Mode**, XamPress keeps everything neatly organized in your home directory under `~/.xampress` (customizable via the `XAMPRESS_HOME` environment variable):
- `~/.xampress/config.json` — Global configuration file
- `~/.xampress/sites/` — All created WordPress and Laravel projects
- `~/.xampress/bin/php/` — Portable PHP CLI binary
- `~/.xampress/bin/mysql/` — Portable MySQL/MariaDB server & client binaries
- `~/.xampress/bin/wp-cli.phar` — WP-CLI executable archive
- `~/.xampress/bin/composer.phar` — Composer executable archive (for Laravel)
- `~/.xampress/mysql-data/` — Local MySQL data directory
- `~/.xampress/run/` — Runtime PID files, router scripts, and server logs

---

## TODO

- [x] Standalone mode with portable PHP, MySQL, WP-CLI & Composer installer
- [x] Built-in PHP server (`php -S`) per-site runner with `-n` (`0.0.0.0`) network flag
- [x] Laravel project creation & validated `xampress artisan` wrapper
- [x] Existing project/site cloning (`xampress clone`)
- [ ] Backup and restore of existing projects (`zip` + SQL dump)
- [ ] One-click / magic login URL generation

---

## Support

For support, contact me on [Telegram](https://t.me/nonesubham) or open an issue on this repository.
