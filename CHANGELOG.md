## 1.9.0 (Unreleased)

UPGRADE NOTES:

NEW FEATURES:

ENHANCEMENTS:
* Added `-show-sensitive` flag to tofu plan, apply, state-show and output commands to display sensitive data in output. ([#1554](https://github.com/opentofu/opentofu/pull/1554))
* Improved performance for large graphs when debug logs are not enabled. ([#1810](https://github.com/opentofu/opentofu/pull/1810))
* Improved performance for large graphs with many submodules. ([#1809](https://github.com/opentofu/opentofu/pull/1809))

BUG FIXES:
* Ensure that using a sensitive path for templatefile that it doesn't panic([#1801](https://github.com/opentofu/opentofu/issues/1801))

## Previous Releases

For information on prior major and minor releases, see their changelogs:

- [v1.8](https://github.com/opentofu/opentofu/blob/v1.8/CHANGELOG.md)
- [v1.7](https://github.com/opentofu/opentofu/blob/v1.7/CHANGELOG.md)
- [v1.6](https://github.com/opentofu/opentofu/blob/v1.6/CHANGELOG.md)
