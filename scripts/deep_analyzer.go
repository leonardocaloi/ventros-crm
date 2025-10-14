package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Analysis results
type AnalysisResult struct {
	// DDD Patterns
	AggregatesWithVersion    []string
	AggregatesWithoutVersion []string
	DomainEvents             map[string][]string // aggregate -> events
	RepositoryInterfaces     []string

	// Clean Architecture violations
	DomainLayerImports   map[string][]string // file -> illegal imports
	DomainDependsOnInfra []string

	// Security issues
	HandlersWithoutTenantCheck []string
	HandlersWithoutAuthCheck   []string
	RawSQLUsage                []string

	// CQRS
	CommandHandlers []string
	QueryHandlers   []string

	// Code metrics
	ComplexFunctions     []ComplexFunction
	LargeFunctions       []FunctionMetric
	PublicFunctionsCount int
	TotalLinesOfCode     int
}

type ComplexFunction struct {
	Name       string
	File       string
	Complexity int
}

type FunctionMetric struct {
	Name  string
	File  string
	Lines int
}

func main() {
	fmt.Println("🔍 Ventros CRM - Deep AST Analysis")
	fmt.Println("=====================================")
	fmt.Println()

	projectRoot := getProjectRoot()
	result := &AnalysisResult{
		DomainEvents:       make(map[string][]string),
		DomainLayerImports: make(map[string][]string),
	}

	// Analyze different layers
	fmt.Println("📦 Analyzing domain layer...")
	analyzeDomainLayer(filepath.Join(projectRoot, "internal/domain"), result)

	fmt.Println("🏗️  Analyzing application layer...")
	analyzeApplicationLayer(filepath.Join(projectRoot, "internal/application"), result)

	fmt.Println("🌐 Analyzing infrastructure layer...")
	analyzeInfrastructureLayer(filepath.Join(projectRoot, "infrastructure"), result)

	// Generate report
	fmt.Println()
	fmt.Println("📊 Generating report...")
	generateReport(result)

	fmt.Println()
	fmt.Println("✅ Deep analysis complete!")
	fmt.Println("📄 Report saved to: DEEP_ANALYSIS_REPORT.md")
}

func getProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// If running from scripts/ directory, go up one level
	if strings.HasSuffix(wd, "scripts") {
		return filepath.Dir(wd)
	}

	return wd
}

func analyzeDomainLayer(domainPath string, result *AnalysisResult) {
	// Walk through domain directory
	filepath.Walk(domainPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil
		}

		// Check for illegal imports (domain should not import infrastructure)
		checkDomainImports(node, path, result)

		// Find aggregates and check for version field
		checkAggregateVersionField(node, path, result)

		// Find domain events
		findDomainEvents(node, path, result)

		// Find repository interfaces
		findRepositoryInterfaces(node, result)

		return nil
	})
}

func checkDomainImports(node *ast.File, filePath string, result *AnalysisResult) {
	illegalPrefixes := []string{
		"github.com/gin-gonic",
		"gorm.io/gorm",
		"github.com/rabbitmq",
		"infrastructure/",
	}

	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)

		for _, prefix := range illegalPrefixes {
			if strings.Contains(importPath, prefix) {
				result.DomainLayerImports[filePath] = append(
					result.DomainLayerImports[filePath],
					importPath,
				)
			}
		}
	}
}

func checkAggregateVersionField(node *ast.File, filePath string, result *AnalysisResult) {
	// Look for struct definitions
	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		// Check if this looks like an aggregate (has ID field)
		hasID := false
		hasVersion := false

		for _, field := range structType.Fields.List {
			if len(field.Names) > 0 {
				fieldName := field.Names[0].Name
				if fieldName == "id" || fieldName == "ID" {
					hasID = true
				}
				if fieldName == "version" || fieldName == "Version" {
					hasVersion = true
				}
			}
		}

		// If it's an aggregate, track version field status
		if hasID {
			aggregateName := getAggregateFromPath(filePath)
			if hasVersion {
				if !contains(result.AggregatesWithVersion, aggregateName) {
					result.AggregatesWithVersion = append(result.AggregatesWithVersion, aggregateName)
				}
			} else {
				if !contains(result.AggregatesWithoutVersion, aggregateName) {
					result.AggregatesWithoutVersion = append(result.AggregatesWithoutVersion, aggregateName)
				}
			}
		}

		return true
	})
}

func findDomainEvents(node *ast.File, filePath string, result *AnalysisResult) {
	// Events are methods that return string and are named EventType()
	ast.Inspect(node, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok || funcDecl.Name.Name != "EventType" {
			return true
		}

		// Get the receiver type (event struct)
		if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
			recvType := funcDecl.Recv.List[0].Type
			eventName := getTypeName(recvType)

			aggregate := getAggregateFromPath(filePath)
			result.DomainEvents[aggregate] = append(result.DomainEvents[aggregate], eventName)
		}

		return true
	})
}

func findRepositoryInterfaces(node *ast.File, result *AnalysisResult) {
	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		// Check if it's an interface with "Repository" in the name
		if _, ok := typeSpec.Type.(*ast.InterfaceType); ok {
			if strings.Contains(typeSpec.Name.Name, "Repository") {
				result.RepositoryInterfaces = append(result.RepositoryInterfaces, typeSpec.Name.Name)
			}
		}

		return true
	})
}

func analyzeApplicationLayer(appPath string, result *AnalysisResult) {
	// Find command and query handlers
	filepath.Walk(appPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil
		}

		// Find command handlers
		if strings.Contains(path, "commands/") {
			findHandlers(node, "Handler", result.CommandHandlers, &result.CommandHandlers)
		}

		// Find query handlers
		if strings.Contains(path, "queries/") {
			findHandlers(node, "Query", result.QueryHandlers, &result.QueryHandlers)
		}

		return nil
	})
}

func findHandlers(node *ast.File, suffix string, existing []string, target *[]string) {
	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		if _, ok := typeSpec.Type.(*ast.StructType); ok {
			if strings.HasSuffix(typeSpec.Name.Name, suffix) {
				*target = append(*target, typeSpec.Name.Name)
			}
		}

		return true
	})
}

func analyzeInfrastructureLayer(infraPath string, result *AnalysisResult) {
	// Analyze HTTP handlers for security issues
	handlersPath := filepath.Join(infraPath, "http/handlers")

	filepath.Walk(handlersPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil
		}

		// Check for security patterns
		checkHandlerSecurity(node, path, result)

		return nil
	})

	// Check for raw SQL usage
	persistencePath := filepath.Join(infraPath, "persistence")
	filepath.Walk(persistencePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		// Check for db.Raw or db.Exec (potential SQL injection)
		if strings.Contains(string(content), "db.Raw(") || strings.Contains(string(content), "db.Exec(") {
			result.RawSQLUsage = append(result.RawSQLUsage, path)
		}

		return nil
	})
}

func checkHandlerSecurity(node *ast.File, filePath string, result *AnalysisResult) {
	// Find handler functions (methods with *gin.Context parameter)
	ast.Inspect(node, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		// Check if function has gin.Context parameter
		hasGinContext := false
		if funcDecl.Type.Params != nil {
			for _, param := range funcDecl.Type.Params.List {
				paramType := getTypeName(param.Type)
				if strings.Contains(paramType, "gin.Context") {
					hasGinContext = true
					break
				}
			}
		}

		if !hasGinContext {
			return true
		}

		// Check function body for security checks
		hasTenantCheck := false
		hasAuthCheck := false

		ast.Inspect(funcDecl.Body, func(bodyNode ast.Node) bool {
			callExpr, ok := bodyNode.(*ast.CallExpr)
			if !ok {
				return true
			}

			// Check for c.GetString("tenant_id")
			if sel, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == "GetString" && len(callExpr.Args) > 0 {
					if lit, ok := callExpr.Args[0].(*ast.BasicLit); ok {
						if strings.Contains(lit.Value, "tenant_id") {
							hasTenantCheck = true
						}
						if strings.Contains(lit.Value, "user_id") || strings.Contains(lit.Value, "auth") {
							hasAuthCheck = true
						}
					}
				}
			}

			return true
		})

		handlerName := fmt.Sprintf("%s (in %s)", funcDecl.Name.Name, filepath.Base(filePath))

		if !hasTenantCheck {
			result.HandlersWithoutTenantCheck = append(result.HandlersWithoutTenantCheck, handlerName)
		}

		if !hasAuthCheck {
			result.HandlersWithoutAuthCheck = append(result.HandlersWithoutAuthCheck, handlerName)
		}

		return true
	})
}

func generateReport(result *AnalysisResult) {
	f, err := os.Create("DEEP_ANALYSIS_REPORT.md")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fmt.Fprintln(f, "# 🔬 VENTROS CRM - DEEP ANALYSIS REPORT (AST-BASED)")
	fmt.Fprintln(f)
	fmt.Fprintln(f, "**Generated**: "+currentTimestamp())
	fmt.Fprintln(f, "**Method**: Go AST parsing + static analysis")
	fmt.Fprintln(f, "**Type**: Deterministic code analysis")
	fmt.Fprintln(f)
	fmt.Fprintln(f, "---")
	fmt.Fprintln(f)

	// 1. DDD Analysis
	fmt.Fprintln(f, "## 1. 🏗️ DOMAIN-DRIVEN DESIGN (DDD)")
	fmt.Fprintln(f)

	// Sort aggregates
	sort.Strings(result.AggregatesWithVersion)
	sort.Strings(result.AggregatesWithoutVersion)

	totalAggregates := len(result.AggregatesWithVersion) + len(result.AggregatesWithoutVersion)
	versionCoverage := 0.0
	if totalAggregates > 0 {
		versionCoverage = float64(len(result.AggregatesWithVersion)) / float64(totalAggregates) * 100
	}

	fmt.Fprintf(f, "### Optimistic Locking Analysis\n\n")
	fmt.Fprintf(f, "| Metric | Count | Coverage |\n")
	fmt.Fprintf(f, "|--------|-------|----------|\n")
	fmt.Fprintf(f, "| Aggregates WITH version field | %d | %.1f%% |\n", len(result.AggregatesWithVersion), versionCoverage)
	fmt.Fprintf(f, "| Aggregates WITHOUT version field | %d | - |\n", len(result.AggregatesWithoutVersion))
	fmt.Fprintf(f, "| **Total Aggregates** | %d | - |\n", totalAggregates)
	fmt.Fprintln(f)

	if len(result.AggregatesWithVersion) > 0 {
		fmt.Fprintln(f, "**✅ Aggregates WITH optimistic locking**:")
		fmt.Fprintln(f)
		for _, agg := range result.AggregatesWithVersion {
			fmt.Fprintf(f, "- ✅ `%s`\n", agg)
		}
		fmt.Fprintln(f)
	}

	if len(result.AggregatesWithoutVersion) > 0 {
		fmt.Fprintln(f, "**🔴 Aggregates WITHOUT optimistic locking (HIGH PRIORITY FIX)**:")
		fmt.Fprintln(f)
		for _, agg := range result.AggregatesWithoutVersion {
			fmt.Fprintf(f, "- 🔴 `%s`\n", agg)
		}
		fmt.Fprintln(f)
		fmt.Fprintln(f, "**Action Required**: Add `version int` field to each aggregate above.")
		fmt.Fprintln(f)
	}

	// Domain Events
	fmt.Fprintln(f, "### Domain Events")
	fmt.Fprintln(f)
	totalEvents := 0
	for _, events := range result.DomainEvents {
		totalEvents += len(events)
	}
	fmt.Fprintf(f, "**Total Domain Events**: %d\n\n", totalEvents)
	fmt.Fprintln(f, "**Events by Aggregate**:")
	fmt.Fprintln(f)

	// Sort aggregates for consistent output
	aggregates := make([]string, 0, len(result.DomainEvents))
	for agg := range result.DomainEvents {
		aggregates = append(aggregates, agg)
	}
	sort.Strings(aggregates)

	for _, agg := range aggregates {
		events := result.DomainEvents[agg]
		fmt.Fprintf(f, "- `%s` (%d events)\n", agg, len(events))
		for _, event := range events {
			fmt.Fprintf(f, "  - `%s`\n", event)
		}
	}
	fmt.Fprintln(f)

	// Repository Interfaces
	fmt.Fprintf(f, "### Repository Interfaces\n\n")
	fmt.Fprintf(f, "**Total Repository Interfaces**: %d\n\n", len(result.RepositoryInterfaces))
	if len(result.RepositoryInterfaces) > 0 {
		sort.Strings(result.RepositoryInterfaces)
		for _, repo := range result.RepositoryInterfaces {
			fmt.Fprintf(f, "- `%s`\n", repo)
		}
	}
	fmt.Fprintln(f)
	fmt.Fprintln(f, "---")
	fmt.Fprintln(f)

	// 2. Clean Architecture
	fmt.Fprintln(f, "## 2. 🎯 CLEAN ARCHITECTURE VIOLATIONS")
	fmt.Fprintln(f)

	if len(result.DomainLayerImports) > 0 {
		fmt.Fprintln(f, "**🔴 CRITICAL: Domain layer has illegal dependencies**")
		fmt.Fprintln(f)
		fmt.Fprintln(f, "Domain layer should NOT import infrastructure frameworks.")
		fmt.Fprintln(f)
		for file, imports := range result.DomainLayerImports {
			relPath := strings.TrimPrefix(file, getProjectRoot()+"/")
			fmt.Fprintf(f, "- `%s`\n", relPath)
			for _, imp := range imports {
				fmt.Fprintf(f, "  - ❌ `%s`\n", imp)
			}
		}
		fmt.Fprintln(f)
	} else {
		fmt.Fprintln(f, "✅ **No Clean Architecture violations detected**")
		fmt.Fprintln(f)
		fmt.Fprintln(f, "Domain layer correctly depends only on itself.")
		fmt.Fprintln(f)
	}

	fmt.Fprintln(f, "---")
	fmt.Fprintln(f)

	// 3. CQRS
	fmt.Fprintln(f, "## 3. 📝 CQRS ANALYSIS")
	fmt.Fprintln(f)
	fmt.Fprintf(f, "| Pattern | Count |\n")
	fmt.Fprintf(f, "|---------|-------|\n")
	fmt.Fprintf(f, "| Command Handlers | %d |\n", len(result.CommandHandlers))
	fmt.Fprintf(f, "| Query Handlers | %d |\n", len(result.QueryHandlers))
	fmt.Fprintln(f)

	fmt.Fprintln(f, "---")
	fmt.Fprintln(f)

	// 4. Security Analysis
	fmt.Fprintln(f, "## 4. 🔒 SECURITY ANALYSIS")
	fmt.Fprintln(f)

	// BOLA
	fmt.Fprintln(f, "### API1:2023 - Broken Object Level Authorization (BOLA)")
	fmt.Fprintln(f)

	if len(result.HandlersWithoutTenantCheck) > 0 {
		fmt.Fprintf(f, "**🔴 %d handlers without tenant_id check**:\n\n", len(result.HandlersWithoutTenantCheck))
		sort.Strings(result.HandlersWithoutTenantCheck)
		for _, handler := range result.HandlersWithoutTenantCheck {
			fmt.Fprintf(f, "- 🔴 `%s`\n", handler)
		}
		fmt.Fprintln(f)
		fmt.Fprintln(f, "**Risk**: Unauthorized access to other tenants' data")
		fmt.Fprintln(f, "**Action**: Add `tenantID := c.GetString(\"tenant_id\")` check")
		fmt.Fprintln(f)
	} else {
		fmt.Fprintln(f, "✅ All handlers have tenant_id checks")
		fmt.Fprintln(f)
	}

	// SQL Injection
	fmt.Fprintln(f, "### SQL Injection Risk")
	fmt.Fprintln(f)

	if len(result.RawSQLUsage) > 0 {
		fmt.Fprintf(f, "**⚠️  %d files use raw SQL (potential risk)**:\n\n", len(result.RawSQLUsage))
		for _, file := range result.RawSQLUsage {
			relPath := strings.TrimPrefix(file, getProjectRoot()+"/")
			fmt.Fprintf(f, "- ⚠️  `%s`\n", relPath)
		}
		fmt.Fprintln(f)
		fmt.Fprintln(f, "**Action**: Ensure all raw SQL uses parameterized queries")
		fmt.Fprintln(f)
	} else {
		fmt.Fprintln(f, "✅ No raw SQL usage detected")
		fmt.Fprintln(f)
	}

	fmt.Fprintln(f, "---")
	fmt.Fprintln(f)

	// 5. Recommendations
	fmt.Fprintln(f, "## 5. 📈 RECOMMENDATIONS (Priority Ordered)")
	fmt.Fprintln(f)

	// Generate prioritized recommendations
	priorities := generatePriorities(result)

	if len(priorities.P0) > 0 {
		fmt.Fprintln(f, "### 🔴 P0 - CRITICAL")
		fmt.Fprintln(f)
		for _, rec := range priorities.P0 {
			fmt.Fprintf(f, "%s\n\n", rec)
		}
	}

	if len(priorities.P1) > 0 {
		fmt.Fprintln(f, "### ⚠️  P1 - HIGH")
		fmt.Fprintln(f)
		for _, rec := range priorities.P1 {
			fmt.Fprintf(f, "%s\n\n", rec)
		}
	}

	if len(priorities.P2) > 0 {
		fmt.Fprintln(f, "### 🟡 P2 - MEDIUM")
		fmt.Fprintln(f)
		for _, rec := range priorities.P2 {
			fmt.Fprintf(f, "%s\n\n", rec)
		}
	}

	fmt.Fprintln(f, "---")
	fmt.Fprintln(f)
	fmt.Fprintln(f, "**End of Deep Analysis Report**")
	fmt.Fprintln(f)
	fmt.Fprintln(f, "*Generated by: scripts/deep_analyzer.go*")
}

type Priorities struct {
	P0 []string
	P1 []string
	P2 []string
}

func generatePriorities(result *AnalysisResult) Priorities {
	p := Priorities{}

	// P0 - Critical
	if len(result.HandlersWithoutTenantCheck) > 0 {
		p.P0 = append(p.P0, fmt.Sprintf(
			"**Fix BOLA vulnerabilities**: %d handlers lack tenant_id checks\n"+
				"   - Impact: CRITICAL - Unauthorized data access\n"+
				"   - Effort: 1-2 weeks",
			len(result.HandlersWithoutTenantCheck),
		))
	}

	if len(result.DomainLayerImports) > 0 {
		p.P0 = append(p.P0, fmt.Sprintf(
			"**Fix Clean Architecture violations**: Domain layer imports infrastructure\n"+
				"   - Files affected: %d\n"+
				"   - Impact: CRITICAL - Architecture integrity\n"+
				"   - Effort: 1 week",
			len(result.DomainLayerImports),
		))
	}

	// P1 - High
	if len(result.AggregatesWithoutVersion) > 0 {
		p.P1 = append(p.P1, fmt.Sprintf(
			"**Add optimistic locking**: %d aggregates missing version field\n"+
				"   - Impact: HIGH - Data corruption risk\n"+
				"   - Effort: 1 day per aggregate",
			len(result.AggregatesWithoutVersion),
		))
	}

	if len(result.RawSQLUsage) > 0 {
		p.P1 = append(p.P1, fmt.Sprintf(
			"**Review raw SQL usage**: %d files use db.Raw/Exec\n"+
				"   - Impact: HIGH - SQL injection risk\n"+
				"   - Effort: 1 week",
			len(result.RawSQLUsage),
		))
	}

	// P2 - Medium
	totalAggregates := len(result.AggregatesWithVersion) + len(result.AggregatesWithoutVersion)
	if totalAggregates > 0 && len(result.AggregatesWithVersion) < totalAggregates {
		coverage := float64(len(result.AggregatesWithVersion)) / float64(totalAggregates) * 100
		if coverage < 100 {
			p.P2 = append(p.P2, fmt.Sprintf(
				"**Complete optimistic locking coverage**: Currently at %.1f%%\n"+
					"   - Target: 100%%\n"+
					"   - Effort: Ongoing",
				coverage,
			))
		}
	}

	return p
}

// Helper functions

func getAggregateFromPath(path string) string {
	parts := strings.Split(path, string(filepath.Separator))
	for i, part := range parts {
		if part == "domain" && i+2 < len(parts) {
			return parts[i+1] + "/" + parts[i+2]
		}
	}
	return "unknown"
}

func getTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + getTypeName(t.X)
	case *ast.SelectorExpr:
		return getTypeName(t.X) + "." + t.Sel.Name
	default:
		return ""
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func currentTimestamp() string {
	return "2025-10-14" // placeholder
}
