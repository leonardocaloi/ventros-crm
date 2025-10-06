# Contributing to Ventros CRM

First off, thank you for considering contributing to Ventros CRM! 🎉

It's people like you that make Ventros CRM such a great tool. We welcome contributions from everyone, whether you're fixing a typo, adding documentation, reporting a bug, or implementing a new feature.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Commit Guidelines](#commit-guidelines)
- [Pull Request Process](#pull-request-process)
- [Community](#community)

---

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to [conduct@ventros.com](mailto:conduct@ventros.com).

**Expected Behavior:**
- Be respectful and inclusive
- Welcome newcomers and help them get started
- Give and receive constructive feedback gracefully
- Focus on what is best for the community
- Show empathy towards other community members

**Unacceptable Behavior:**
- Harassment, discrimination, or offensive comments
- Trolling, insulting/derogatory comments
- Public or private harassment
- Publishing others' private information without permission
- Other conduct which could reasonably be considered inappropriate

---

## Getting Started

### Prerequisites

Before you begin, ensure you have:
- Go 1.25.1 or later
- Docker & Docker Compose
- Git
- A GitHub account

### Fork & Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/ventros-crm.git
   cd ventros-crm
   ```
3. Add upstream remote:
   ```bash
   git remote add upstream https://github.com/caloi/ventros-crm.git
   ```

### Set Up Development Environment

```bash
# Copy environment variables
cp .env.example .env

# Start infrastructure
make infra-up

# Run migrations
make migrate

# Seed data
make seed

# Run tests to verify setup
make test
```

---

## How Can I Contribute?

### Reporting Bugs 🐛

Before creating bug reports, please check the [issue tracker](https://github.com/caloi/ventros-crm/issues) to avoid duplicates.

**When filing a bug report, include:**
- **Clear title** - Describe the issue concisely
- **Steps to reproduce** - Detailed steps to reproduce the behavior
- **Expected behavior** - What you expected to happen
- **Actual behavior** - What actually happened
- **Environment** - OS, Go version, Docker version, etc.
- **Logs** - Relevant log output (use code blocks)
- **Screenshots** - If applicable

**Template:**
```markdown
## Bug Description
A clear and concise description of what the bug is.

## Steps to Reproduce
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

## Expected Behavior
What you expected to happen.

## Actual Behavior
What actually happened.

## Environment
- OS: [e.g. Ubuntu 22.04]
- Go Version: [e.g. 1.25.1]
- Docker Version: [e.g. 24.0.0]

## Logs
```
paste logs here
```

## Screenshots
If applicable, add screenshots.
```

### Suggesting Enhancements 💡

Enhancement suggestions are tracked as [GitHub issues](https://github.com/caloi/ventros-crm/issues).

**When suggesting an enhancement, include:**
- **Clear title** - Describe the enhancement concisely
- **Use case** - Explain why this would be useful
- **Proposed solution** - Describe how you envision it working
- **Alternatives** - Other approaches you've considered
- **Additional context** - Screenshots, mockups, links, etc.

### Improving Documentation 📚

Documentation improvements are always welcome! This includes:
- Fixing typos or grammatical errors
- Adding examples or clarifications
- Creating new guides or tutorials
- Improving code comments
- Translating documentation

### Writing Code 💻

Look for issues tagged with:
- `good first issue` - Good for newcomers
- `help wanted` - Extra attention needed
- `bug` - Something isn't working
- `enhancement` - New feature or request

---

## Development Workflow

### 1. Create a Branch

Always create a new branch for your work:
```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

**Branch naming conventions:**
- `feature/` - New features (e.g., `feature/add-email-channel`)
- `fix/` - Bug fixes (e.g., `fix/session-timeout`)
- `docs/` - Documentation (e.g., `docs/update-readme`)
- `refactor/` - Code refactoring (e.g., `refactor/extract-use-case`)
- `test/` - Adding tests (e.g., `test/contact-repository`)
- `chore/` - Maintenance (e.g., `chore/update-dependencies`)

### 2. Make Your Changes

- Write clean, readable code
- Follow the [Coding Standards](#coding-standards)
- Add tests for new features
- Update documentation as needed

### 3. Test Your Changes

```bash
# Run all tests
make test

# Run specific tests
go test ./internal/domain/contact/...

# Run with coverage
make test-coverage

# Run linters
make lint
```

### 4. Commit Your Changes

Follow our [Commit Guidelines](#commit-guidelines):
```bash
git add .
git commit -m "feat: add email channel integration"
```

### 5. Push & Create Pull Request

```bash
# Push to your fork
git push origin feature/your-feature-name

# Open a Pull Request on GitHub
```

---

## Coding Standards

### Go Code Style

We follow standard Go conventions:
- Use `gofmt` for formatting (run automatically with `make lint`)
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use [golangci-lint](https://golangci-lint.run/) (run with `make lint`)

### Architecture Principles

#### Domain-Driven Design (DDD)
- **Aggregates** should be self-contained and enforce invariants
- **Value Objects** should be immutable
- **Domain Events** should be raised for important state changes
- **Repositories** should only work with Aggregate Roots

#### SOLID Principles
- **Single Responsibility** - Each struct/function should have one reason to change
- **Open/Closed** - Open for extension, closed for modification
- **Liskov Substitution** - Subtypes must be substitutable
- **Interface Segregation** - Many specific interfaces > one general
- **Dependency Inversion** - Depend on abstractions, not concretions

#### Clean Architecture
```
Handlers → Use Cases → Domain
   ↓           ↓
Infrastructure ← Repositories
```

### File Organization

```
/cmd               - Main applications (entry points)
/internal          - Private application code
  /domain          - Domain layer (Aggregates, Entities, Value Objects)
  /application     - Application layer (Use Cases, DTOs)
  /workflows       - Temporal workflows
/infrastructure    - Infrastructure layer (DB, HTTP, Messaging)
/guides            - Documentation
/scripts           - Build/deploy scripts
/deployments       - Deployment configs (Docker, Helm)
```

### Naming Conventions

- **Packages**: lowercase, singular, no underscores (e.g., `contact`, `session`)
- **Files**: lowercase, underscores for separation (e.g., `contact_repository.go`)
- **Types**: PascalCase (e.g., `Contact`, `ContactRepository`)
- **Functions**: PascalCase for exported, camelCase for unexported
- **Variables**: camelCase (e.g., `contactID`, `sessionTimeout`)
- **Constants**: PascalCase or ALL_CAPS for exports

### Comments

- Use complete sentences with proper punctuation
- Comment exported functions, types, and constants
- Add godoc comments for public APIs
- Explain **why**, not **what** (code should be self-explanatory)

```go
// Good
// NewContact creates a new Contact aggregate with validation.
// Returns an error if projectID or tenantID are invalid.
func NewContact(projectID, tenantID, name string) (*Contact, error) {
    // ...
}

// Bad
// NewContact creates a contact
func NewContact(projectID, tenantID, name string) (*Contact, error) {
    // ...
}
```

### Error Handling

- Return errors, don't panic (except for truly exceptional cases)
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Use custom error types for domain errors
- Log errors at the appropriate level

```go
// Good
if err := repo.Save(ctx, contact); err != nil {
    return fmt.Errorf("failed to save contact %s: %w", contact.ID(), err)
}

// Bad
if err := repo.Save(ctx, contact); err != nil {
    panic(err) // Don't panic!
}
```

### Testing

- Write table-driven tests
- Use meaningful test names (TestFunctionName_Scenario_ExpectedResult)
- Mock external dependencies
- Aim for >70% code coverage

```go
func TestContact_SetEmail_ValidEmail_Success(t *testing.T) {
    // Arrange
    contact, _ := NewContact(projectID, tenantID, "John")
    
    // Act
    err := contact.SetEmail("john@example.com")
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "john@example.com", contact.Email().String())
}
```

---

## Commit Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/):

### Format
```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types
- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation changes
- **style**: Code style changes (formatting, missing semicolons, etc.)
- **refactor**: Code refactoring
- **test**: Adding or updating tests
- **chore**: Maintenance tasks (dependencies, build, etc.)
- **perf**: Performance improvements
- **ci**: CI/CD changes

### Scope (optional)
- `contact` - Contact domain
- `session` - Session domain
- `message` - Message domain
- `api` - API changes
- `db` - Database changes
- `helm` - Helm chart changes

### Examples
```bash
feat(contact): add email validation with regex
fix(session): resolve timeout not being extended
docs(readme): update installation instructions
refactor(handlers): extract use cases from handlers
test(contact): add unit tests for value objects
chore(deps): update Go to 1.25.1
```

### Rules
- Use imperative, present tense ("add" not "added" or "adds")
- Don't capitalize first letter
- No period at the end
- Limit subject to 72 characters
- Reference issues in footer (e.g., "Closes #123")

---

## Pull Request Process

### Before Submitting

- [ ] Code follows project style guidelines
- [ ] Tests pass locally (`make test`)
- [ ] Linters pass (`make lint`)
- [ ] Documentation is updated
- [ ] CHANGELOG.md is updated (if applicable)
- [ ] Commits follow commit guidelines
- [ ] Branch is up to date with `main`

### PR Template

When opening a PR, use this template:

```markdown
## Description
Brief description of what this PR does.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Related Issue
Closes #(issue number)

## How Has This Been Tested?
Describe the tests you ran to verify your changes.

## Checklist
- [ ] My code follows the style guidelines
- [ ] I have performed a self-review of my code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have updated the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] CHANGELOG.md is updated

## Screenshots (if applicable)
Add screenshots to help explain your changes.
```

### Review Process

1. **Automated checks** must pass (CI/CD)
2. **Code review** by at least one maintainer
3. **Approval** from maintainer
4. **Merge** by maintainer (squash & merge)

### After Merge

- Your contribution will be included in the next release
- You'll be added to the contributors list
- Close any related issues

---

## Community

### Communication Channels

- **GitHub Issues** - Bug reports and feature requests
- **GitHub Discussions** - General questions and discussions
- **Email** - [dev@ventros.com](mailto:dev@ventros.com)

### Getting Help

If you need help:
1. Check the [documentation](guides/)
2. Search [existing issues](https://github.com/caloi/ventros-crm/issues)
3. Ask in [GitHub Discussions](https://github.com/caloi/ventros-crm/discussions)
4. Reach out to maintainers

### Recognition

Contributors are recognized in:
- README.md contributors section
- Release notes
- GitHub contributors page

---

## Questions?

Don't hesitate to ask! We're here to help. Open an issue or start a discussion.

**Thank you for contributing! 🚀**
