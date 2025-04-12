# 🛠 dev-setup

A blazing-fast CLI tool to **bootstrap microservice development environments** with a single command.  
Ideal for platform and dev experience (DevEx) teams that want to reduce onboarding time and eliminate “works on my machine” issues.

---

## 🚀 What It Does

`dev-setup` automates the typical multi-service dev environment workflow:

1. ✅ **Clones** all relevant repositories in parallel  
2. 🐳 **Runs docker-compose** or docker containers per service  
3. 🔧 **Injects environment variables** and shared `.env` values  
4. ❤️ **Runs pre/post hooks** for migrations, builds, setup steps  
5. 🩺 **Performs health checks** (HTTP, command, TCP)  
6. 🎯 **Declares "ready"** once all services are up and healthy

---

## 📦 Features

- 🧠 Declarative config (`dev-setup.yaml`)
- 🧩 Supports Docker & Docker Compose
- 🔁 Automatic repo discovery
- 🧵 Pre/Post hook execution (e.g., run migrations after boot)
- 🔍 Built-in health check engine
- 💨 Parallel execution with smart orchestration

---

## 🛠 Example `dev-setup.yaml`

```yaml
globalEnv:
  NODE_ENV: development

services:
  - name: user-service
    repo: git@github.com:acme/user-service.git
    path: ./user-service
    compose-file: docker-compose.yml
    hooks:
      - name: migrate
        trigger: post-up
        run: "npm run migrate"

  - name: auth-service
    repo: git@github.com:acme/auth-service.git
    path: ./auth-service
    compose-file: docker-compose.yml
    healthCheck:
      type: http
      url: http://localhost:4000/health
```
--- 

## Use cases
Onboarding developers to large mono/multi-repo setups

Bootstrapping internal tools quickly

Aligning local and production service environments

Running platform team health checks and hooks