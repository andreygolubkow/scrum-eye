Dev Digest — Daily Console Digest for Scrum Masters

Overview
- A console tool that aggregates daily insights from your development tooling:
  - TeamCity builds and test statistics for selected build configurations and branch.
  - Azure Boards: current sprint features and progress.
  - Azure Repos: repositories and branches matching feature IDs from Azure Boards.
- Colorful, human-friendly output designed for quick reading in the terminal.
- Modular architecture with discoverable modules based on community patterns.

Install
1) Ensure Go 1.21+ is installed.
2) Clone this repository.
3) Run: go build -o dev-digest

Configuration
- Create a config.yaml in the project root (or set DEV_DIGEST_CONFIG env var to a file path). Tokens can be inlined or referenced as ${ENV_VAR}.

Example config.yaml

console:
  # color: true|false (default auto)
  timeout: 20s # default per-module timeout

teamcity:
  base_url: https://teamcity.example.com
  token: ${TEAMCITY_TOKEN}
  branch: develop
  builds: [ "MyApp_Build", "MyLib_Build" ]
  timeout: 25s

azure:
  organization: https://dev.azure.com/your-org
  pat: ${AZURE_PAT}
  timeout: 25s
  boards:
    project: YourProject
    team: YourTeam
  repos: {}

Tokens and permissions
- TeamCity: create a token with permissions to read builds and tests.
- Azure DevOps PAT: scopes required
  - Work Items (Read) to query Boards
  - Code (Read) to read Repos

Usage
- Build: go build -o dev-digest
- Run: ./dev-digest -config ./config.yaml
- Disable colors: ./dev-digest -no-color

Design
- common package: shared interfaces and models (Module, Config, Report).
- console package: orchestrates loading config, running modules with timeouts, and rendering.
- modules/* packages: individual modules self-register via init() using common.Register.

Notes
- All HTTP calls respect timeouts and are best-effort. Errors from a module are shown in its section without failing the whole run.
- If a module section in the config is omitted, the module is disabled automatically.
- Azure Repos module uses Feature IDs collected from Azure Boards to search for matching branches across repositories.

Limitations
- API schemas can change; endpoints used are minimal and may need adjustments for your environment.
- The tool avoids heavy dependencies and focuses on a clear, readable output.
