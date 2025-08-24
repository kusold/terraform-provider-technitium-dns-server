# Terraform Provider for Technitium DNS Server - Development Plan

## Project Overview

Create a comprehensive Terraform provider for managing Technitium DNS Server instances, following HashiCorp provider design principles and enabling Infrastructure as Code for DNS management.

## 🚀 Progress Summary

**Status**: Phase 2 Complete ✅, Ready for Phase 3
**Last Updated**: August 2025

### Recently Completed (August 2025)

- ✅ **Project Foundation**: Go module initialization, directory structure, and basic provider skeleton
- ✅ **Development Environment**: Nix flakes with direnv integration for reproducible builds
- ✅ **Code Quality**: golangci-lint configuration with Terraform provider standards
- ✅ **Task Automation**: Comprehensive Taskfile.yaml with full development workflow
- ✅ **Testing Infrastructure**: Complete TestContainers integration with parallel execution
- ✅ **Provider Framework**: Basic provider implementation with Terraform Plugin Framework v1.15.1
- ✅ **Test Suite**: Working unit tests, acceptance framework, and comprehensive fixtures
- ✅ **Example Configurations**: Complete resource, data source, and usage examples
- ✅ **Parallel Testing**: ParallelTestRunner with container lifecycle management
- ✅ **Acceptance Testing**: Full framework with provider factories and test helpers
- ✅ **Provider Core Implementation**: Complete provider schema, API client, authentication, and zone management foundation
- ✅ **Primary Zone Resource**: Complete `technitium_zone` resource with full CRUD operations, comprehensive schema, and testing
- ✅ **DNS Apps Implementation**: Complete DNS Apps resource and data sources with full testing
- ✅ **DNS Apps Testing**: Complete integration, acceptance, and configuration management tests

### Current State

- ✅ Code compiles without errors
- ✅ Dependencies resolved and locked
- ✅ Complete example configurations for all implemented resources
- ✅ Production-ready testing infrastructure
- ✅ **Provider Core Completed**: Full API client with authentication, error handling, retry logic, and zone management
- ✅ **Primary Zone Resource Completed**: Full CRUD operations, import support, comprehensive schema with all zone types
- ✅ **DNS Record Resource Completed**: Full CRUD operations for all major DNS record types
- ✅ **DNS Apps Implementation Completed**: Complete resource and data sources with comprehensive testing
- ✅ **Unit Tests**: All Phase 1 and Phase 2 resource and client unit tests passing
- ✅ **DNS Apps Testing**: Complete integration, acceptance, and configuration management tests

### Current Phase Status

- **Phase 1 Complete** ✅
  - ✅ Implemented `technitium_dns_record` resource
  - ✅ Implemented `technitium_zone` data source
  - ✅ Implemented `technitium_dns_records` data source
  - ✅ Completed comprehensive documentation
- **Phase 2 Complete** ✅
  - ✅ Implemented `technitium_dns_app` resource (fully tested)
  - ✅ Implemented `technitium_dns_apps` data source (fully tested)
  - ✅ Implemented `technitium_dns_store_apps` data source (fully tested)
  - ✅ Complete DNS Apps API client implementation (9/9 methods with unit tests)
  - ✅ Comprehensive unit tests for DNS Apps resource and data sources
  - ✅ Integration tests for DNS Apps functionality (all 4 tests completed)
  - ✅ Acceptance tests for DNS Apps functionality (full lifecycle testing)
  - ✅ Testing of app configuration management functionality
  - ✅ Testing of file upload functionality for app installation
- **Next**: Begin Phase 3 - Enhanced DNS Management
  - [ ] Implement advanced zone types (secondary, forwarder)
  - [ ] Extend DNS record types (SSHFP, TLSA, CAA, DS)
  - [ ] Implement DNSSEC management

## Repository Setup & Infrastructure

### Initial Project Setup

- [x] Initialize Go module with proper naming
- [x] Set up directory structure following Terraform provider conventions
- [x] Configure Nix flakes for reproducible development environment
- [x] Set up direnv integration
- [ ] Configure Git hooks and pre-commit automation
- [ ] Set up GitHub repository with proper templates and workflows

### Development Infrastructure

- [x] Configure golangci-lint with Terraform provider standards
- [x] Set up Taskfile.yaml with comprehensive development workflow
- [ ] Configure VS Code/IDE settings for Go development
- [ ] Set up debugging configuration for provider development

### Testing Infrastructure

- [x] Set up TestContainers integration with Technitium Docker image
- [x] Configure acceptance testing framework
- [x] Set up test helper utilities and common test data
- [x] Create test fixtures and example configurations
- [x] Set up parallel test execution configuration

## Phase 1: Foundation (Core Zone Management) - Weeks 1-2

### Provider Core

- [x] Implement provider schema with configuration options:
  - [x] Server URL (required)
  - [x] Username/password authentication
  - [x] API token authentication
  - [x] Timeout and retry configurations
  - [x] TLS/SSL settings
- [x] Create Technitium API client with:
  - [x] HTTP client with proper timeouts
  - [x] Authentication handling (session vs API tokens)
  - [x] Error mapping from API responses to Terraform diagnostics
  - [x] Rate limiting and retry logic
  - [x] Request/response logging for debugging

### Primary Zone Resource (`technitium_zone`)

- [x] Implement resource schema:
  - [x] Zone name (required)
  - [x] Zone type (primary, secondary, forwarder, etc.)
  - [x] Catalog zone membership
  - [x] SOA serial date scheme configuration
  - [x] Primary name server addresses (for secondary zones)
  - [x] Zone transfer protocol settings (TCP, TLS, QUIC)
  - [x] TSIG key authentication
  - [x] ZONEMD validation settings
  - [x] Forwarder initialization and configuration
  - [x] DNS transport protocols
  - [x] DNSSEC validation settings
  - [x] Proxy configuration (type, address, port, credentials)
  - [x] Computed attributes (internal, dnssec_status, disabled, soa_serial)
- [x] Implement CRUD operations:
  - [x] Create zone with validation and comprehensive parameter support
  - [x] Read zone configuration and state via zones/options/get API
  - [x] Update zone settings via zones/options/set API
  - [x] Delete zone with proper cleanup via zones/delete API
- [x] Add import functionality for existing zones
- [x] Implement proper state management and drift detection
- [x] Add comprehensive unit tests and schema validation
- [x] Register resource with provider

### Basic DNS Records Resource (`technitium_dns_record`) ✅

- [x] Implement resource schema for basic record types:
  - [x] A records (IPv4 addresses)
  - [x] AAAA records (IPv6 addresses)
  - [x] CNAME records (canonical names)
  - [x] MX records (mail exchange)
  - [x] TXT records (text data)
  - [x] NS records (name server)
  - [x] PTR records (reverse DNS)
  - [x] SRV records (service location)
- [x] Implement CRUD operations:
  - [x] Create records with type-specific validation
  - [x] Read record data and metadata
  - [x] Update existing records
  - [x] Delete records with dependency checks
- [x] Add TTL management and validation
- [x] Implement record import functionality

### Data Sources

- [x] Implement `technitium_zone` data source:
  - [x] Query existing zone information
  - [x] Return comprehensive zone metadata
- [x] Implement `technitium_dns_records` data source:
  - [x] List all records in a zone
  - [x] Support filtering by record type
  - [x] Return record details and metadata

### Testing & Documentation

- [x] Write unit tests for zone resource
- [x] Write unit tests for DNS record resource
- [x] Create integration tests using TestContainers
- [x] Write acceptance tests for core functionality
- [x] Generate initial provider documentation
- [x] Create basic usage examples
- [x] Document authentication methods and provider configuration

## Phase 2: DNS Apps Management - Weeks 3-4 ✅

### DNS Apps Resource (`technitium_dns_app`) ✅ Complete

- [x] Implement resource schema:
  - [x] App name (required)
  - [x] App version
  - [x] Install method (url, file upload)
  - [x] App URL for download and install/update
  - [x] App configuration data
  - [x] Computed attributes (installed_version, dns_apps metadata)
- [x] Implement CRUD operations:
  - [x] Create app via download and install from URL
  - [x] Create app via file upload (multipart form data)
  - [x] Read app information and configuration
  - [x] Update app via download and update from URL
  - [x] Update app via file upload
  - [x] Delete app (uninstall) with proper cleanup
- [x] Add import functionality for existing apps
- [x] Implement app configuration management (get/set config)
- [x] Add comprehensive unit tests and schema validation
- [x] Register resource with provider

### DNS Apps Data Sources ✅ Complete

- [x] Implement `technitium_dns_apps` data source:
  - [x] List all installed apps
  - [x] Return app metadata and DNS app details
  - [x] Support filtering by app type/characteristics
- [x] Implement `technitium_dns_store_apps` data source:
  - [x] List all available store apps
  - [x] Return app store metadata
  - [x] Include installation status and update availability

### DNS Apps API Client Methods ✅ Complete

- [x] Implement client methods for all DNS Apps API calls:
  - [x] `ListApps()` - List installed apps
  - [x] `ListStoreApps()` - List store apps
  - [x] `DownloadAndInstallApp()` - Install from URL
  - [x] `DownloadAndUpdateApp()` - Update from URL
  - [x] `InstallApp()` - Install via file upload
  - [x] `UpdateApp()` - Update via file upload
  - [x] `UninstallApp()` - Remove app
  - [x] `GetAppConfig()` - Retrieve app configuration
  - [x] `SetAppConfig()` - Save app configuration
- [x] Add proper error handling and validation
- [x] Implement file upload support for multipart form data
- [x] Add retry logic for download operations

### Testing & Documentation ✅ Complete

- [x] Write unit tests for all 9 DNS apps client methods
- [x] Write unit tests for DNS apps resource CRUD operations (schema validation, configure method)
- [x] Write unit tests for DNS apps data sources (schema validation, configure method)
- [x] Create integration tests using TestContainers (4 comprehensive test files)
- [x] Write acceptance tests for app lifecycle management
- [x] Test app configuration management functionality
- [x] Test file upload functionality with mock ZIP files
- [x] Document DNS apps usage patterns
- [x] Create examples for common app installation scenarios

## Phase 3: Enhanced DNS Management - Weeks 5-6

### Advanced Zone Types

- [ ] Implement `technitium_secondary_zone` resource:
  - [ ] Primary name server configuration
  - [ ] Zone transfer protocol settings (TCP, TLS, QUIC)
  - [ ] TSIG key authentication
  - [ ] Zone validation options
- [ ] Implement `technitium_forwarder_zone` resource:
  - [ ] Conditional forwarder configuration
  - [ ] Forwarder addresses and priorities
  - [ ] Protocol selection (UDP, TCP, TLS, HTTPS, QUIC)
  - [ ] DNSSEC validation settings
  - [ ] Proxy configuration

### Extended DNS Record Types

- [ ] Extend `technitium_dns_record` with additional types:
  - [ ] NS records (name server delegation)
  - [ ] SRV records (service location)
  - [ ] PTR records (reverse DNS)
  - [ ] SSHFP records (SSH fingerprints)
  - [ ] TLSA records (TLS associations)
  - [ ] CAA records (certificate authority authorization)
  - [ ] DS records (delegation signer)
- [ ] Add specialized validation for each record type
- [ ] Implement proper dependency handling between record types

### DNSSEC Management

- [ ] Implement `technitium_dnssec` resource:
  - [ ] Zone signing configuration
  - [ ] Algorithm selection (RSA, ECDSA, EDDSA)
  - [ ] Key rollover policies
  - [ ] NSEC/NSEC3 configuration
  - [ ] DS record management for parent zones
- [ ] Add DNSSEC status monitoring and validation
- [ ] Implement key lifecycle management

### Enhanced Testing & Documentation

- [ ] Expand test coverage for all Phase 2 features
- [ ] Add complex scenario tests (multi-zone setups)
- [ ] Create DNSSEC-specific test scenarios
- [ ] Update documentation with advanced DNS features
- [ ] Add troubleshooting guides

## Phase 4: Access Control & Configuration - Weeks 7-8

### Zone Access Control

- [ ] Implement `technitium_zone_permissions` resource:
  - [ ] User-based permissions (view, modify, delete)
  - [ ] Group-based permissions
  - [ ] Zone-specific access control
  - [ ] Permission inheritance handling

### Allow/Block List Management

- [ ] Implement `technitium_allowed_zone` resource:
  - [ ] Domain allowlist management
  - [ ] Bulk import/export functionality
  - [ ] Pattern-based rules
- [ ] Implement `technitium_blocked_zone` resource:
  - [ ] Domain blocklist management
  - [ ] Integration with external blocklist feeds
  - [ ] Custom blocking responses

### DNS Server Configuration

- [ ] Implement `technitium_dns_settings` resource:
  - [ ] Server domain and endpoints configuration
  - [ ] Recursion and forwarding settings
  - [ ] Cache configuration and limits
  - [ ] Logging and monitoring settings
  - [ ] Security and rate limiting options
  - [ ] Protocol-specific settings (DNS-over-HTTPS, DNS-over-TLS, etc.)

### User & Group Management

- [ ] Implement `technitium_user` resource:
  - [ ] User account creation and management
  - [ ] Password and session policies
  - [ ] Group membership management
- [ ] Implement `technitium_group` resource:
  - [ ] Group creation and management
  - [ ] Permission assignments
  - [ ] Member management

### Testing & Documentation

- [ ] Create comprehensive access control tests
- [ ] Test permission inheritance and conflicts
- [ ] Add configuration validation tests
- [ ] Document security best practices
- [ ] Create multi-user scenario examples

## Phase 5: Advanced Features - Weeks 9-10

### DHCP Integration

- [ ] Implement `technitium_dhcp_scope` resource:
  - [ ] DHCP scope configuration
  - [ ] IP address range management
  - [ ] DHCP options configuration
  - [ ] Lease management
  - [ ] Integration with DNS records

### TSIG Keys

- [ ] Implement `technitium_tsig_key` resource:
  - [ ] TSIG key generation and management
  - [ ] Algorithm support
  - [ ] Key rotation policies
  - [ ] Zone transfer authentication

### Additional Data Sources

- [ ] Implement `technitium_dns_settings` data source
- [ ] Implement `technitium_stats` data source for monitoring
- [ ] Implement `technitium_users` and `technitium_groups` data sources

## Quality Assurance & Polish

### Code Quality

- [ ] Achieve >90% test coverage across all packages
- [ ] Implement comprehensive error handling
- [ ] Add detailed logging and debugging support
- [ ] Optimize API calls and reduce unnecessary requests
- [ ] Implement proper resource timeouts

### Documentation & Examples

- [ ] Complete provider documentation with all resources and data sources
- [x] Create comprehensive examples for common use cases
- [ ] Write migration guides from manual management
- [ ] Document best practices and security considerations
- [ ] Create video tutorials or demos

### Performance & Reliability

- [ ] Implement request batching where possible
- [ ] Add circuit breaker patterns for API reliability
- [ ] Optimize state refresh operations
- [ ] Implement proper resource dependencies
- [ ] Add provider-level configuration validation

## Release & Distribution

### CI/CD Pipeline

- [ ] Set up GitHub Actions for automated testing
- [ ] Configure multi-platform builds (Linux, macOS, Windows)
- [ ] Set up automated security scanning
- [ ] Configure release automation with proper versioning
- [ ] Set up integration testing with multiple Technitium versions

### Release Preparation

- [ ] Prepare provider for Terraform Registry submission
- [ ] Create comprehensive README with installation instructions
- [ ] Prepare CHANGELOG.md with semantic versioning
- [ ] Set up GPG signing for releases
- [ ] Create provider verification and validation

### Community & Maintenance

- [ ] Set up issue and PR templates
- [ ] Create contribution guidelines
- [ ] Set up community support channels
- [ ] Plan maintenance and update schedule
- [ ] Create backward compatibility policy

## Technical Debt & Future Enhancements

### Potential Future Features

- [ ] Terraform Cloud/Enterprise integration
- [ ] Provider configuration via environment variables
- [ ] Bulk operations and import utilities
- [ ] Monitoring and alerting integrations
- [ ] Backup and restore functionality
- [ ] Multi-server provider support

### Performance Optimizations

- [ ] Implement provider schema caching
- [ ] Add connection pooling for API clients
- [ ] Optimize large zone handling
- [ ] Implement incremental state updates

## Success Metrics

- [ ] Provider successfully manages DNS zones and records
- [ ] Comprehensive test coverage with reliable CI/CD
- [ ] Published to Terraform Registry with positive community feedback
- [ ] Documentation rated as comprehensive and helpful
- [ ] Performance benchmarks meet acceptable thresholds
- [ ] Security review completed with no critical issues

---

## Notes for Implementation

- Follow HashiCorp provider design principles throughout development
- Ensure each resource represents a single API object
- Maintain close alignment between resource schema and Technitium API
- Prioritize importability for all resources
- Use Terraform Plugin Framework v6 for modern development patterns
- Ensure proper state management and drift detection
- Focus on user experience and clear error messages
