# terraform-provider-technitium-dns-server
This project is a terraform provider for controlling the Technitium DNS Server.

@TODO.md contains all the tasks that need to accomplished in this project.

## Technitium DNS Server
Technitium is a an open source DNS Server.

The Technitium API Docs are located at .ai/docs/technitium-api.md. Ensure you refer to it whenever you have questions about how the API works, or what the request/response body looks like.

Other useful links are below:
* [GitHub Repository](https://github.com/TechnitiumSoftware/DnsServer)
* [Published Docker Image](https://hub.docker.com/r/technitium/dns-server)
* [Docker Environment Variable Docs](https://raw.githubusercontent.com/TechnitiumSoftware/DnsServer/refs/heads/master/DockerEnvironmentVariables.md)

## Common Commands
This project utilizes [Taskfile](https://taskfile.dev) as the task runner. You can find all available commands in Taskfile.yaml.

## Code Style Guidelines
See @.ai/docs/hashicorp-provider-design-principles.mdx for terraform provider design principles.

## Testing Instructions
Unit tests are the preference, however I don't want you to do perform excessive mocking.

Because this project tightly integrates with Technitium's API, it's important that we are able to validate our implementation is correct, and that we can detect regressions as Technitium updates. This project utilizes [TestContainers for Go](https://golang.testcontainers.org) to spin up the Technitium docker image to test against.

## Developer Environment
This project uses Nix flakes to ensure that all the required dependencies are available. It should also utilize direnv to automatically initialize the nix shell.

## Workflow
Whenever you start a new task, it should be performed in a branch.

- The project's plan is stored in TODO.md. Update it as you complete tasks