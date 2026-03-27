// Package cache provides SQLite-based persistence and caching for AST nodes,
// relationships, violations, and file metadata.
//
// The cache package implements:
//   - ASTCache: SQLite database for AST node storage
//   - Incremental analysis via file hash comparison
//   - Dual-pool database connections (read/write separation)
//   - Git-aware caching and branch tracking
//   - Database migrations and schema versioning
//   - Violation caching and querying
//
// # Database Location
//
// Cache database is stored in:
//
//	~/.cache/arch-unit/ast.db
//
// Database structure:
//   - ast_nodes: Code elements (packages, types, methods, fields)
//   - ast_relationships: Dependencies between nodes
//   - library_nodes: External library references
//   - library_relationships: Calls to external libraries
//   - file_metadata: File hashes and timestamps
//   - violations: Architecture rule violations
//   - dependency_aliases: Package name aliases
//
// # Core Functionality
//
// ## AST Caching
//
// Store and retrieve AST nodes with automatic de-duplication:
//
//	cache, _ := cache.GetASTCache()
//	defer cache.Close()
//
//	// Store AST node
//	node := &models.ASTNode{
//	    FilePath:    "user.go",
//	    PackageName: "services",
//	    TypeName:    "UserService",
//	    MethodName:  "GetUser",
//	    NodeType:    models.NodeTypeMethod,
//	    StartLine:   42,
//	}
//	nodeID, _ := cache.StoreASTNode(node)
//
//	// Retrieve nodes by file
//	nodes, _ := cache.GetASTNodesByFile("user.go")
//
// ## Incremental Analysis
//
// Skip unchanged files using file hash comparison:
//
//	needsAnalysis, _ := cache.NeedsReanalysis("user.go")
//	if !needsAnalysis {
//	    // Use cached data
//	    nodes, _ := cache.GetASTNodesByFile("user.go")
//	} else {
//	    // Re-analyze and update cache
//	    result, _ := analyzer.AnalyzeFile(task, "user.go", content)
//	    cache.UpdateFileMetadata("user.go")
//	}
//
// File hash is calculated using SHA256:
//
//	hash := sha256(file_content)
//	metadata.FileHash = hex(hash)
//
// ## Relationship Storage
//
// Track dependencies between code elements:
//
//	// Store method call relationship
//	cache.StoreASTRelationship(
//	    fromID,                // Caller node ID
//	    &toID,                 // Called node ID
//	    lineNo,                // Line number
//	    "call",                // Relationship type
//	    "repo.FindByID(id)",   // Call text
//	)
//
//	// Retrieve relationships
//	relationships, _ := cache.GetASTRelationships(nodeID, "call")
//
// Relationship types:
//   - call: Method/function calls
//   - reference: Variable/type references
//   - inheritance: Type inheritance
//   - implements: Interface implementation
//   - import: Package imports
//
// ## Library Tracking
//
// Track calls to external libraries:
//
//	// Store library node
//	libraryID, _ := cache.StoreLibraryNode(
//	    "fmt",                 // Package
//	    "",                    // Class (empty for Go)
//	    "Println",             // Method
//	    "",                    // Field
//	    models.NodeTypeMethod, // Node type
//	    "go",                  // Language
//	    "stdlib",              // Framework
//	)
//
//	// Store library relationship
//	cache.StoreLibraryRelationship(nodeID, libraryID, lineNo, "call", "fmt.Println")
//
// # Dual-Pool Database
//
// Separate read and write connections for performance:
//
//	// Read operations (uses read-only pool)
//	readDB := cache.GetReadQuery()
//	var nodes []*models.ASTNode
//	readDB.Where("file_path = ?", file).Find(&nodes)
//
//	// Write operations (uses write pool)
//	writeDB := cache.GetWriteQuery()
//	writeDB.Create(&node)
//
// Benefits:
//   - Concurrent reads without blocking
//   - Write serialization prevents corruption
//   - WAL mode for better concurrency
//   - Read-only protection on read pool
//
// # Singleton Pattern
//
// Cache uses singleton pattern for shared access:
//
//	cache1, _ := cache.GetASTCache()
//	cache2, _ := cache.GetASTCache()
//	// cache1 == cache2 (same instance)
//
// Reset singleton for testing:
//
//	cache.ResetASTCache()
//	newCache, _ := cache.GetASTCache()
//
// # Update-First Strategy
//
// Preserve node IDs across re-analysis:
//
//	// First analysis
//	node1 := &models.ASTNode{PackageName: "services", TypeName: "UserService"}
//	id1, _ := cache.StoreASTNode(node1)  // ID: 42
//
//	// Re-analysis (file modified)
//	node2 := &models.ASTNode{PackageName: "services", TypeName: "UserService"}
//	id2, _ := cache.StoreASTNode(node2)  // ID: 42 (same!)
//
// Benefits:
//   - Stable node IDs for relationships
//   - Efficient updates (no cascade deletes)
//   - Preserves relationship integrity
//
// # Git Integration
//
// Track AST data by git branch:
//
//	gitCache := cache.NewGitCache()
//	defer gitCache.Close()
//
//	// Store data for current branch
//	gitCache.SetBranchData("main", data)
//
//	// Retrieve data for branch
//	data, _ := gitCache.GetBranchData("main")
//
// Supports:
//   - Branch-specific caching
//   - Automatic branch detection
//   - Cross-branch comparison
//
// # Violation Caching
//
// Store and query architecture violations:
//
//	violation := &models.Violation{
//	    File:     "user.go",
//	    Line:     42,
//	    CallerID: &callerID,
//	    CalledID: &calledID,
//	    Source:   "arch-unit",
//	}
//	cache.StoreViolation(violation)
//
//	// Query violations
//	violations, _ := cache.GetViolations(
//	    cache.WithFile("user.go"),
//	    cache.WithSource("arch-unit"),
//	)
//
// # Database Migrations
//
// Automatic schema migrations on startup:
//
//	migrationManager, _ := cache.NewMigrationManager()
//	defer migrationManager.Close()
//
//	err := migrationManager.RunMigrations()
//
// Migration features:
//   - Version tracking
//   - Automatic schema updates
//   - Rollback support
//   - Migration history
//
// # Transaction Support
//
// GORM transactions for consistency:
//
//	err := cache.GetWriteQuery().Transaction(func(tx *gorm.DB) error {
//	    // Store multiple nodes atomically
//	    for _, node := range nodes {
//	        if err := tx.Create(node).Error; err != nil {
//	            return err // Rollback
//	        }
//	    }
//	    return nil // Commit
//	})
//
// # Performance Optimization
//
// ## Indexes
//
// Database indexes for fast queries:
//   - file_path: File-based node lookup
//   - package_name: Package-based queries
//   - type_name, method_name: Code element search
//   - node_type: Filter by element type
//   - start_line: Line-based lookups
//
// ## Batch Operations
//
// StoreFileResults for efficient bulk inserts:
//
//	cache.StoreFileResults("user.go", &types.ASTResult{
//	    Nodes:         nodes,
//	    Relationships: relationships,
//	    Libraries:     libraries,
//	})
//
// Single transaction for entire file analysis.
//
// ## Connection Pooling
//
// Configurable connection pools:
//
//	db := newDualPoolGormDB()
//	// Read pool: 10 connections
//	// Write pool: 1 connection
//
// # Virtual Paths
//
// Support for non-file sources:
//
//	// SQL connection
//	cache.UpdateFileMetadata("sql://localhost:5432/mydb")
//
//	// OpenAPI spec
//	cache.UpdateFileMetadata("openapi://https://api.example.com/spec")
//
//	// Generic virtual
//	cache.UpdateFileMetadata("virtual://config/settings")
//
// Virtual paths use path-based hash instead of file content hash.
//
// # Cleanup Operations
//
// Clear cache data:
//
//	// Clear all data
//	cache.ClearAllData()
//
//	// Clear data for specific file
//	cache.DeleteASTForFile("user.go")
//
//	// Remove old entries
//	cache.CleanOldEntries(30 * 24 * time.Hour)  // 30 days
//
// # Testing Support
//
// In-memory database for tests:
//
//	cache, _ := cache.NewASTCacheWithPath(":memory:")
//	defer cache.Close()
//	// Fast tests without disk I/O
//
// # Line-Based Queries
//
// Find nodes by line number:
//
//	node := cache.FindByLine("user.go", 42)
//	// Returns most specific node at line 42
//
// Useful for:
//   - Jump to definition
//   - Code navigation
//   - Error location mapping
//
// # Error Handling
//
// Graceful handling of common issues:
//   - Database locked: Retry with backoff
//   - File not found: Return nil, no error
//   - Corrupted data: Clear and rebuild
//   - Schema mismatch: Run migrations
//
// # Thread Safety
//
// Cache is thread-safe with:
//   - GORM connection pooling
//   - WAL mode for concurrent access
//   - Mutex protection for critical sections
//   - Read-only connections for reads
//
// See also:
//   - github.com/flanksource/gavel/models for data structures
//   - github.com/flanksource/gavel/analysis for AST extraction
//   - gorm.io/gorm for database operations
package cache
