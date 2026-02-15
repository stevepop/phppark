# PHPark

A modern development environment manager for Linux - inspired by Laravel Valet and Herd.

Manage multiple PHP applications with automatic Nginx configuration, instant PHP version switching, and `.test` domain support - all from a simple CLI.

[![Release](https://img.shields.io/github/v/release/stevepop/phppark)](https://github.com/stevepop/phppark/releases)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Features

- 🚀 **One-Command Setup**: Install everything (nginx, dnsmasq, PHP) with `phppark setup`
- 🐘 **Multiple PHP Versions**: Auto-install and switch between PHP 7.4, 8.0, 8.1, 8.2, 8.3, 8.4
- ⚡ **Instant Switching**: Changes PHP version immediately for both sites and CLI
- 🔧 **Auto-Deploy**: Automatic Nginx config generation and deployment
- 🌐 **`.test` Domains**: Automatic local domain resolution
- 🔒 **SSL Support**: Self-signed certificates for HTTPS development
- 🔐 **Auto-Permissions**: Automatic permission fixes for served directories
- 📦 **Zero Configuration**: Just install and start building
- ⚙️ **Service Management**: Auto-start and manage nginx and PHP-FPM

## Why PHPark?

PHPark eliminates the manual setup traditionally required for PHP development on Linux. No more editing nginx configs, managing PHP-FPM pools, or wrestling with file permissions. Just park your projects and start coding.

## Quick Start

### Installation

**On Ubuntu/Debian:**
```bash
# Download latest release
wget https://github.com/stevepop/phppark/releases/latest/download/phppark-linux
chmod +x phppark-linux
sudo mv phppark-linux /usr/local/bin/phppark

# One-command setup (installs nginx, dnsmasq, PHP 8.2)
sudo phppark setup
```

### Create Your First Site
```bash
# Create a site
mkdir -p ~/sites/myapp/public
echo '<?php phpinfo();' > ~/sites/myapp/public/index.php

# Park it (auto-configures everything)
cd ~/sites
sudo phppark park

# Set up DNS so .test domains resolve
sudo phppark trust

# Done! Your site is running at myapp.test
curl http://myapp.test
```

## Commands

### Site Management
```bash
phppark park [path]          # Serve all subdirectories as sites
phppark link [name]          # Link current directory as a site
phppark unlink [name]        # Remove a site
phppark links                # List all sites
phppark rebuild              # Rebuild all nginx configs
```

### PHP Version Management
```bash
phppark use 8.3              # Switch PHP version globally (sites + CLI)
phppark use 8.2 mysite       # Switch PHP version for specific site
phppark php:list             # List available PHP versions
```

**PHPark automatically installs any PHP version you request!** No manual setup needed.

### SSL
```bash
phppark secure [site]        # Add HTTPS to site
phppark unsecure [site]      # Remove HTTPS from site
```

### DNS
```bash
phppark trust                # Setup DNS resolution for .test domains
phppark untrust              # Remove DNS configuration
```

### System
```bash
phppark status               # Show PHPark configuration and system info
phppark install              # Initialize PHPark configuration
phppark setup                # Complete system setup (recommended)
```

## Complete Workflow Example
```bash
# 1. One-time setup (on fresh Ubuntu)
sudo phppark setup

# 2. Set up DNS so .test domains resolve
sudo phppark trust

# 3. Create your first app
mkdir -p ~/sites/blog/public
echo '<?php phpinfo();' > ~/sites/blog/public/index.php

# 4. Park the sites directory
cd ~/sites
sudo phppark park

# 5. Your site is ready at blog.test!
curl http://blog.test

# 6. Need a different PHP version?
sudo phppark use 8.3         # Switches globally (CLI + sites)
sudo phppark rebuild         # Apply to existing sites

# 7. Create another app with specific PHP version
mkdir -p ~/sites/api/public
cd ~/sites/api
sudo phppark link api
sudo phppark use 8.2 api     # This site uses PHP 8.2

# 8. Add HTTPS
sudo phppark secure blog

# Your sites:
# - https://blog.test (PHP 8.3)
# - http://api.test (PHP 8.2)
```

## How It Works

PHPark automates your entire PHP development environment:

1. **Automatic Nginx Configuration**: Generates and deploys nginx configs with optimal settings
2. **Permission Management**: Fixes directory permissions automatically so nginx can serve your files
3. **Service Management**: Starts and restarts nginx and PHP-FPM as needed
4. **Smart PHP Installation**: Detects missing PHP versions and installs them on demand
5. **Instant CLI Switching**: Updates your system PHP CLI version immediately
6. **DNS Resolution**: Configures dnsmasq for seamless `.test` domain routing

**Zero manual configuration. Just works.**

## What Makes PHPark Different

- **Truly Zero-Config**: From bare Ubuntu to working sites in under 2 minutes
- **Intelligent Automation**: Auto-installs dependencies, fixes permissions, manages services
- **Production-Grade Configs**: Optimized nginx configurations for Laravel, Symfony, and all PHP frameworks
- **Developer-Friendly**: Helpful error messages, clear status reporting, intuitive commands
- **Built for Linux**: Native Linux tool, not a port - fast and lightweight

## Requirements

- Ubuntu 20.04+ or Debian-based Linux
- `sudo` access (for nginx/service management)

That's it! PHPark installs everything else automatically.

## Manual Installation (Advanced)

If you prefer to install dependencies manually:
```bash
# Install dependencies
sudo apt update
sudo apt install -y nginx dnsmasq software-properties-common
sudo add-apt-repository -y ppa:ondrej/php
sudo apt install -y php8.2-fpm

# Install PHPark
wget https://github.com/stevepop/phppark/releases/latest/download/phppark-linux
chmod +x phppark-linux
sudo mv phppark-linux /usr/local/bin/phppark

# Initialize
sudo phppark install
```

## Troubleshooting

### Sites not accessible
```bash
# Check status
sudo phppark status

# Rebuild configs
sudo phppark rebuild

# Verify nginx
sudo nginx -t
sudo systemctl status nginx
```

### PHP version not switching
```bash
# Check available versions
phppark php:list

# Verify PHP-FPM is running
sudo systemctl status php8.2-fpm

# Check site config
sudo cat /etc/nginx/sites-enabled/mysite.conf
```

## Configuration

PHPark stores its configuration in `~/.phppark/` (or `/root/.phppark/` when using sudo):

- `config.yaml` - Main configuration
- `sites.json` - Registered sites
- `nginx/` - Generated nginx configs
- `certificates/` - SSL certificates

Edit `config.yaml` to customize:
```yaml
domain: .test        # Change to .local, .dev, etc.
defaultPHP: "8.2"   # Default PHP version
https: false        # Enable HTTPS by default
```

## Development Status

**v1.0.0 - Production Ready** ✅

All core features implemented and production-validated:
- [x] Complete system setup automation
- [x] Automatic nginx configuration and deployment
- [x] Multi-version PHP support with auto-installation
- [x] Instant CLI PHP version switching
- [x] SSL certificate generation and management
- [x] DNS resolution via dnsmasq
- [x] Automatic permission management
- [x] Service orchestration (nginx, PHP-FPM)
- [x] Site parking and linking
- [x] Per-site PHP version control

## Roadmap

**v1.1.0** - Enhanced Features
- Database management (MySQL/PostgreSQL)
- Redis/Memcached support
- Custom domain support (beyond `.test`)
- Site templates/scaffolding
- Backup and restore functionality

**v2.0.0** - Advanced Features
- Docker integration
- Multi-distro support (Arch, Fedora, etc.)
- Cloud deployment helpers
- Team collaboration features
- Web-based control panel

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

PHPark was inspired by Laravel Valet and Herd, bringing their elegant development experience to Linux.

## Author

**Steve Popoola** - [stevepop](https://github.com/stevepop)

---

Built with ❤️ for the Linux/PHP community

**⭐ If PHPark makes your development easier, please star the repo!**