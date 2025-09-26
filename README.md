# My Workflow Project
A workflow management project built with Go, featuring MongoDB integration, structured logging, and configuration management.


## Environment Requirements
- Go 1.23.2
- MongoDB (required for data storage)
- Redis (planned for caching)
- SQLite (alternative storage option, planned)

## Quick Start
1. Clone the Repository
```bash
git clone https://github.com/huangpiaopiao7290/my_workflow.git
cd my_workflow
```
2. Install Dependencies
```bash
go mod tidy
```
3. Configure the Project  
> Create a configuration file in the config directory (supports .yaml, .yml, .toml formats). The default configuration path is config/config.yaml:
```bash
# Example config/config.yaml
mongodb:
  username: your_username
  password: your_password
  addr: "mongodb://%s:%s@localhost:27017/"
  database: workflow_db
  maxPoolSize: 100
  minPoolSize: 10
  maxConnIdleTime: 300
  connectionTimeout: 10
```
4. run the application
```bash
go run cmd/server/main.go # -params if you need
```

## Project Structure
- `cmd/server/main.go`: Entry point of the application
- `pkg/database/mongodb`: MongoDB client implementation with singleton pattern and connection pooling
- `internal/module/card`: Data models for workflow cards and attachments
- `config`: Configuration management using Viper (see config/analyze.go for implementation)
- `pkg/logger`: Structured logging with Zerolog, supporting daily log rotation
- `pkg/common/types`: Common type definitions

## Notes
- Add sensitive configuration (credentials, secrets) to .env (ignored by git)
- IDE-specific files (.vscode, .idea) are ignored
- Log files in storage/logs/ are not committed to version control
- Configuration files in config/ with extensions .yaml, .yml, .toml are ignored by git (store sample configs separately)# My Workflow Project  

