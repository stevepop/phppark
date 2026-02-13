# PHPark

A modern development environment manager for Linux - inspired by Laravel Valet and Herd.

Easily manage multiple PHP applications with different PHP versions, automatic Nginx configuration, and `.test` domain support.

> ⚠️ **Early Development**: PHPark is under active development and not yet ready for production use.

## Features (Planned)

- 🐘 **Multiple PHP Versions**: Run PHP 7.4, 8.0, 8.1, 8.2, 8.3 side-by-side
- 🔧 **Automatic Nginx Configuration**: Zero-config site serving
- 🌐 **`.test` Domains**: Automatic local domain resolution via dnsmasq
- 🔒 **SSL Support**: Self-signed certificates for HTTPS development
- ⚡ **Fast**: Written in Go for speed and portability
- 📦 **Simple CLI**: Intuitive commands for site management

## Why PHPark?

Laravel Valet is Mac-only, and Herd doesn't support Linux. PHPark brings the same simple, fast development environment experience to Linux developers.

## Planned Commands
```bash
# Site Management
phppark park [path]          # Serve all subdirectories as sites
phppark link [name]          # Link current directory as a site
phppark unlink [name]        # Remove a site
phppark links                # List all linked sites

# PHP Version Management
phppark use php8.2           # Switch PHP version globally
phppark use php8.2 mysite    # Switch PHP version for specific site
phppark php-list             # List available PHP versions

# SSL
phppark secure [site]        # Add HTTPS to site
phppark unsecure [site]      # Remove HTTPS from site

# Service Management
phppark start                # Start all services
phppark stop                 # Stop all services
phppark restart              # Restart all services
```

## Requirements

- Ubuntu 20.04+ (other distros coming soon)
- Nginx
- PHP-FPM (one or more versions)
- dnsmasq
- openssl

## Development Status

- [x] Project initialization
- [x] CLI framework setup
- [x] Nginx configuration management
- [x] PHP version detection and management
- [x] DNS configuration (dnsmasq)
- [x] SSL certificate generation
- [x] Site linking and parking
- [ ] Service management (nginx reload/restart)

## Contributing

PHPark is open source and contributions are welcome! This project is just getting started.

## License

MIT License - see [LICENSE](LICENSE)

## Roadmap

**v0.1.0 - MVP**
- Basic park/link/unlink functionality
- Single PHP version support
- HTTP only (no SSL)

**v0.2.0 - Multi-PHP**
- Multiple PHP version support
- Per-site PHP version switching

**v0.3.0 - SSL & Polish**
- HTTPS support
- Improved error handling
- Better documentation

**v1.0.0 - Production Ready**
- Full feature parity with Valet
- Comprehensive testing
- Installation packages

## Author

**Steve Popoola** - [stevepop](https://github.com/stevepop)

---

Built with ❤️ for the Linux/PHP community