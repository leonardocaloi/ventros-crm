#!/bin/bash
# Ventros CRM - Test Discovery
# Parseia testes do código e gera comandos dinamicamente
# Usado por: Makefile (make test.*)

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# ============================================
# Descoberta de Testes
# ============================================

discover_tests() {
    local category="$1"  # unit, integration, e2e
    local filter="$2"    # Opcional: waha, domain, message, etc

    case "$category" in
        "unit")
            base_path="internal"
            pattern="*_test.go"
            exclude="tests/"
            ;;
        "integration")
            base_path="tests/integration"
            pattern="*_test.go"
            exclude=""
            ;;
        "e2e")
            base_path="tests/e2e"
            pattern="*_test.go"
            exclude=""
            ;;
        *)
            echo -e "${YELLOW}Unknown category: $category${NC}"
            echo "Usage: discover.sh {unit|integration|e2e} [filter]"
            exit 1
            ;;
    esac

    # Encontra todos os testes
    if [ -n "$filter" ]; then
        # Com filtro
        tests=$(find "$base_path" -type f -name "$pattern" -path "*$filter*" 2>/dev/null | sort)
    else
        # Sem filtro
        tests=$(find "$base_path" -type f -name "$pattern" 2>/dev/null | sort)
    fi

    # Exclui testes de integração/e2e dos unit tests
    if [ "$category" = "unit" ] && [ -n "$exclude" ]; then
        tests=$(echo "$tests" | grep -v "$exclude" || true)
    fi

    # Agrupa por package
    packages=$(echo "$tests" | sed 's|/[^/]*\.go$||' | sort -u)

    echo "$packages"
}

# ============================================
# Análise de Testes
# ============================================

analyze_test_file() {
    local test_file="$1"

    # Extrai funções de teste
    test_functions=$(grep -o "func Test[A-Za-z_]*" "$test_file" | sed 's/func //' || true)

    # Extrai tabelas de teste (table-driven tests)
    table_tests=$(grep -o "name:.*\".*\"" "$test_file" | cut -d'"' -f2 || true)

    # Extrai benchmarks
    benchmarks=$(grep -o "func Benchmark[A-Za-z_]*" "$test_file" | sed 's/func //' || true)

    echo "File: $test_file"
    if [ -n "$test_functions" ]; then
        echo "  Tests: $(echo "$test_functions" | wc -l)"
        echo "$test_functions" | sed 's/^/    - /'
    fi
    if [ -n "$table_tests" ]; then
        echo "  Table tests: $(echo "$table_tests" | wc -l)"
    fi
    if [ -n "$benchmarks" ]; then
        echo "  Benchmarks: $(echo "$benchmarks" | wc -l)"
    fi
    echo ""
}

# ============================================
# Geração de Comandos
# ============================================

generate_makefile_targets() {
    local category="$1"

    echo "# Auto-generated test targets for: $category"
    echo ""

    packages=$(discover_tests "$category" "")

    # Target principal
    echo "test.$category: ## Run all $category tests"
    echo -e "\t@./scripts/make/test/run.sh $category"
    echo ""

    # Sub-targets baseados em packages
    for package in $packages; do
        # Extrai nome do package
        package_name=$(basename "$package")

        # Sanitiza nome para Makefile
        target_name=$(echo "$package_name" | tr '_' '-')

        echo "test.$category.$target_name: ## Run $category tests for $package_name"
        echo -e "\t@./scripts/make/test/run.sh $category $package_name"
        echo ""
    done
}

# ============================================
# Estatísticas
# ============================================

print_stats() {
    local category="$1"

    echo -e "${GREEN}=== Test Statistics: $category ===${NC}"

    packages=$(discover_tests "$category" "")

    total_packages=$(echo "$packages" | wc -l)
    echo "Packages: $total_packages"

    total_files=0
    for package in $packages; do
        files=$(find "$package" -maxdepth 1 -name "*_test.go" 2>/dev/null | wc -l)
        total_files=$((total_files + files))
    done
    echo "Test files: $total_files"

    # Conta funções de teste
    if [ "$category" = "unit" ]; then
        base="internal"
    else
        base="tests/$category"
    fi

    total_tests=$(find "$base" -name "*_test.go" -exec grep -h "func Test" {} \; 2>/dev/null | wc -l)
    echo "Test functions: $total_tests"

    total_benchmarks=$(find "$base" -name "*_test.go" -exec grep -h "func Benchmark" {} \; 2>/dev/null | wc -l)
    if [ "$total_benchmarks" -gt 0 ]; then
        echo "Benchmarks: $total_benchmarks"
    fi

    echo ""
}

# ============================================
# Main
# ============================================

main() {
    local action="${1:-help}"
    shift || true

    case "$action" in
        "list")
            category="$1"
            filter="${2:-}"
            discover_tests "$category" "$filter"
            ;;

        "analyze")
            test_file="$1"
            analyze_test_file "$test_file"
            ;;

        "generate")
            category="$1"
            generate_makefile_targets "$category"
            ;;

        "stats")
            category="${1:-all}"
            if [ "$category" = "all" ]; then
                print_stats "unit"
                print_stats "integration"
                print_stats "e2e"
            else
                print_stats "$category"
            fi
            ;;

        "help"|*)
            echo "Ventros CRM - Test Discovery Tool"
            echo ""
            echo "Usage:"
            echo "  discover.sh list {unit|integration|e2e} [filter]"
            echo "  discover.sh analyze <test_file>"
            echo "  discover.sh generate {unit|integration|e2e}"
            echo "  discover.sh stats [category]"
            echo ""
            echo "Examples:"
            echo "  # List all unit test packages"
            echo "  discover.sh list unit"
            echo ""
            echo "  # List unit tests for domain"
            echo "  discover.sh list unit domain"
            echo ""
            echo "  # Analyze specific test file"
            echo "  discover.sh analyze internal/domain/contact/contact_test.go"
            echo ""
            echo "  # Generate Makefile targets for e2e tests"
            echo "  discover.sh generate e2e"
            echo ""
            echo "  # Show test statistics"
            echo "  discover.sh stats"
            ;;
    esac
}

main "$@"
