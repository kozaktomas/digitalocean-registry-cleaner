# DigitalOcean Registry Cleaner

A CLI tool to automatically clean up old and unused Docker images in your DigitalOcean Container Registry.

## Overview

The DigitalOcean Registry Cleaner (`dorc`) helps you manage container image tags in your DigitalOcean Container Registry
by automatically removing outdated tags while preserving:

- Protected tags (like `latest`, `main`, `master`, `prod`, `production`)
- The most recent N release tags
- Recently updated branch tags (within a specified minimum age)

## Features

- 🧹 **Automatic Cleanup**: Remove old tags based on configurable retention policies
- 🛡️ **Protected Tags**: Safeguard important tags from deletion (main, master, prod, production, latest)
- 🏷️ **Smart Tag Detection**: Distinguishes between release tags (semantic/calendar/sequential versioning) and branch
  tags
- 📅 **Age-Based Filtering**: Keep recent tags and only delete older ones
- 🔒 **Keep Latest**: Retain a specified number of the most recent release tags
- 🔍 **Dry Run Mode**: Preview what would be deleted without making actual changes
- 📊 **Multiple Repositories**: Clean up multiple repositories in a single run

## Usage

```bash
export DO_TOKEN=<your-digitalocean-token> # registry access required
$ ./dorc run --help
Command deletes tags older than [min-age-days] in the registry except the last [keep-tags] tags per repository.

Usage:
  dorc run [flags]

Flags:
      --dry-run                  Dry run
  -h, --help                     help for run
      --keep-tags int            How many tags to keep per repository (default 5)
      --min-age-days int         Minimum age of the tags to delete in days (default 30)
      --protect stringArray      Protect tag/branch (default [latest,main,master,prod,production])
      --registry string          Registry name
      --repository stringArray   Repository name
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
