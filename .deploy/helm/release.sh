#!/usr/bin/env bash

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_header() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  🚀 Ventros CRM Helm Chart Release Script${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

usage() {
    cat << EOF
Usage: $0 <version> [message]

Arguments:
  version    Version to release (e.g., 0.1.0, 1.2.3)
  message    Optional release message (default: "Release version X.Y.Z")

Examples:
  $0 0.1.0
  $0 0.1.0 "Initial release with PostgreSQL operator"
  $0 1.2.3 "Add Redis clustering support"

Notes:
  - Version should follow Semantic Versioning (MAJOR.MINOR.PATCH)
  - Script will create and push a git tag (v<version>)
  - GitHub Actions will automatically publish the Helm chart
  - Ensure you're on the main branch and have no uncommitted changes

EOF
    exit 1
}

check_prerequisites() {
    print_info "Checking prerequisites..."

    # Check if git is installed
    if ! command -v git &> /dev/null; then
        print_error "git is not installed"
        exit 1
    fi

    # Check if helm is installed
    if ! command -v helm &> /dev/null; then
        print_warning "helm is not installed (optional, but recommended for testing)"
    fi

    # Check if we're in a git repository
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        print_error "Not in a git repository"
        exit 1
    fi

    # Check if we're on main branch
    CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
    if [ "$CURRENT_BRANCH" != "main" ] && [ "$CURRENT_BRANCH" != "master" ]; then
        print_warning "You're not on main/master branch (current: $CURRENT_BRANCH)"
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi

    # Check for uncommitted changes
    if ! git diff-index --quiet HEAD --; then
        print_error "You have uncommitted changes. Please commit or stash them first."
        git status --short
        exit 1
    fi

    # Check if remote is configured
    if ! git remote get-url origin &> /dev/null; then
        print_error "No remote 'origin' configured"
        exit 1
    fi

    print_success "All prerequisites met"
}

validate_version() {
    local version=$1
    
    # Check if version follows semantic versioning (X.Y.Z)
    if ! [[ $version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        print_error "Invalid version format: $version"
        print_info "Version must follow Semantic Versioning (e.g., 0.1.0, 1.2.3)"
        exit 1
    fi
    
    print_success "Version format is valid: $version"
}

check_version_exists() {
    local version=$1
    local chart_file="$(git rev-parse --show-toplevel)/.deploy/helm/ventros-crm/Chart.yaml"
    
    if [ ! -f "$chart_file" ]; then
        print_error "Chart.yaml not found"
        exit 1
    fi
    
    local current_version=$(grep "^version:" "$chart_file" | awk '{print $2}')
    
    if [ "$current_version" == "$version" ]; then
        print_error "Chart already at version $version"
        print_info "Current version in Chart.yaml: $current_version"
        print_info "Please use a different version number"
        exit 1
    fi
    
    print_success "Version $version is available (current: $current_version)"
}

lint_chart() {
    print_info "Linting Helm chart..."
    
    local chart_dir="$(git rev-parse --show-toplevel)/.deploy/helm/ventros-crm"
    
    if [ ! -d "$chart_dir" ]; then
        print_error "Chart directory not found: $chart_dir"
        exit 1
    fi
    
    if command -v helm &> /dev/null; then
        if helm lint "$chart_dir"; then
            print_success "Helm chart lint passed"
        else
            print_error "Helm chart lint failed"
            exit 1
        fi
    else
        print_warning "Skipping helm lint (helm not installed)"
    fi
}

update_chart_version() {
    local version=$1
    local chart_file="$2/Chart.yaml"
    
    print_info "Updating Chart.yaml version to $version..."
    
    if [ ! -f "$chart_file" ]; then
        print_error "Chart.yaml not found at: $chart_file"
        exit 1
    fi
    
    # Update version in Chart.yaml
    sed -i "s/^version:.*/version: $version/" "$chart_file"
    sed -i "s/^appVersion:.*/appVersion: \"$version\"/" "$chart_file"
    
    print_success "Chart.yaml updated"
}

create_release() {
    local version=$1
    local message=$2
    local chart_dir="$(git rev-parse --show-toplevel)/.deploy/helm/ventros-crm"
    
    print_info "Creating release for version $version..."
    
    # Update Chart.yaml
    update_chart_version "$version" "$chart_dir"
    
    # Stage changes
    if git add "$chart_dir/Chart.yaml"; then
        print_success "Chart.yaml staged"
    else
        print_error "Failed to stage Chart.yaml"
        exit 1
    fi
    
    # Commit changes
    if git commit -m "chore: bump chart version to $version

$message"; then
        print_success "Changes committed"
    else
        print_error "Failed to commit changes"
        exit 1
    fi
    
    # Push to main
    print_info "Pushing to main branch..."
    if git push origin main; then
        print_success "Changes pushed to main"
    else
        print_error "Failed to push to main"
        print_warning "You can manually push with: git push origin main"
        exit 1
    fi
}

print_next_steps() {
    local version=$1
    local repo_url=$(git remote get-url origin | sed 's/\.git$//')
    local repo_name=$(basename "$repo_url")
    local repo_owner=$(basename "$(dirname "$repo_url")")
    
    # Extract GitHub username from URL
    if [[ $repo_url =~ github\.com[:/]([^/]+)/([^/]+) ]]; then
        repo_owner="${BASH_REMATCH[1]}"
        repo_name="${BASH_REMATCH[2]}"
    fi
    
    echo ""
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}  🎉 Release Created Successfully!${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    print_info "Version: ${version}"
    print_info "Changes pushed to: origin/main"
    echo ""
    print_info "Next steps:"
    echo ""
    echo "  1. 🔄 GitHub Actions is now detecting the version change"
    echo "     Monitor: ${repo_url}/actions"
    echo ""
    echo "  2. ⏱️  Wait ~2-3 minutes for the workflow to complete"
    echo ""
    echo "  3. 📦 The action will automatically:"
    echo "     - Package the Helm chart"
    echo "     - Create GitHub Release (v${version})"
    echo "     - Publish to GitHub Pages"
    echo "     - Update index.yaml"
    echo ""
    echo "  4. 🌐 Your chart will be available at:"
    echo "     https://${repo_owner}.github.io/${repo_name}/"
    echo ""
    echo "  5. 🚀 Users can install with:"
    echo ""
    echo "     helm repo add ventros https://${repo_owner}.github.io/${repo_name}/"
    echo "     helm repo update"
    echo "     helm install ventros-crm ventros/ventros-crm --version ${version}"
    echo ""
    echo "  6. 📋 View the release (after workflow completes):"
    echo "     ${repo_url}/releases/tag/v${version}"
    echo ""
    print_success "Release process initiated!"
    echo ""
}

# Main script
main() {
    print_header
    
    # Parse arguments
    if [ $# -lt 1 ]; then
        usage
    fi
    
    VERSION=$1
    MESSAGE=${2:-"Release version ${VERSION}"}
    TAG="v${VERSION}"
    
    # Run checks
    check_prerequisites
    validate_version "$VERSION"
    check_version_exists "$VERSION"
    lint_chart
    
    # Confirm release
    echo ""
    print_warning "About to create release:"
    echo "  Version: $VERSION"
    echo "  Message: $MESSAGE"
    echo ""
    read -p "Continue? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Release cancelled"
        exit 0
    fi
    
    # Create and push release
    create_release "$VERSION" "$MESSAGE"
    
    # Print next steps
    print_next_steps "$VERSION"
}

# Run main function
main "$@"
