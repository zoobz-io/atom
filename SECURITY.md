# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| Latest  | :white_check_mark: |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue, please follow these steps:

1. **DO NOT** create a public GitHub issue
2. Email security details to the maintainers
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if available)

## Security Best Practices

When using Atom:

1. **Validation**: Always implement the `Validator` interface to validate data before atomization. This prevents malformed data from entering your storage layer.

2. **ID Security**: The `Atoms.ID` field is used as a key identifier. Ensure IDs don't contain sensitive information and are properly sanitized.

3. **Encoding**: The encoding utilities are designed for internal use. When exposing data externally, consider additional encryption or encoding as needed.

4. **Field Exposure**: Be aware that field names and types are discoverable via metadata. Review your struct definitions for information disclosure.

## Security Features

Atom is designed with security in mind:

- Minimal dependencies (only sentinel)
- No network operations
- No file system operations
- Validation hooks before atomization
- Type-safe generics prevent runtime type errors

## Acknowledgments

We appreciate responsible disclosure of security vulnerabilities.
