# v0.6.0

ENHANCEMENTS

* check: Add `-ignore-file-mismatch-data-sources` option
* check: Add `-ignore-file-mismatch-resources` option
* check: Add `-ignore-file-missing-data-sources` option
* check: Add `-ignore-file-missing-resources` option

# v0.5.3

BUG FIXES

* check: Prevent additional errors when `docs/` contains files outside Terraform Provider documentation

# v0.5.2

BUG FIXES

* check: Prevent `mixed Terraform Provider documentation directory layouts found` error when using `website/docs` and `docs/` contains files outside Terraform Provider documentation

# v0.5.1

Released without changes.

# v0.5.0

ENHANCEMENTS

* check: Verify sidebar navigation for missing links and mismatched link text (if legacy directory structure)

# v0.4.1

BUG FIXES

* check: Only verify valid file extensions at end of path (e.g. support additional periods in guide paths) (#25)

# v0.4.0

ENHANCEMENTS

* check: Accept newline-separated files of allowed subcategories with `-allowed-guide-subcategories-file` and `-allowed-resource-subcategories-file` flags
* check: Improve readability with allowed subcategories values in allowed subcategories frontmatter error

# v0.3.0

ENHANCEMENTS

* check: Verify deprecated `sidebar_current` frontmatter is not present

# v0.2.0

ENHANCEMENTS

* check: Verify number of documentation files for Terraform Registry storage limits
* check: Verify size of documentation files for Terraform Registry storage limits
* check: Verify all known data sources and resources have an associated documentation file (if `-providers-schema-json` is provided)
* check: Verify no extraneous or incorrectly named documentation files exist (if `-providers-schema-json` is provided)

# v0.1.2

BUG FIXES

* Remove extraneous `-''` from version information

# v0.1.1

BUG FIXES

* Fix help formatting of `check` command options

# v0.1.0

FEATURES

* Initial release with `check` command
