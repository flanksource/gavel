export interface Test {
  name: string;
  package?: string;
  package_path?: string;
  suite?: string[];
  message?: string;
  file?: string;
  line?: number;
  framework?: string;
  duration?: number; // nanoseconds
  skipped?: boolean;
  failed?: boolean;
  passed?: boolean;
  pending?: boolean;
  stdout?: string;
  stderr?: string;
  children?: Test[];
  summary?: TestSummary;
  context?: GoTestContext | GinkgoContext | FixtureContext;
}

export interface TestSummary {
  Total: number;
  Passed: number;
  Failed: number;
  Skipped: number;
  Pending: number;
  Duration: number;
}

export interface Snapshot {
  tests: Test[];
  done: boolean;
}

export interface GoTestContext {
  parent_test?: string;
  import_path?: string;
}

export interface GinkgoContext {
  suite_description?: string;
  suite_path?: string;
  failure_location?: string;
}

export interface FixtureContext {
  command?: string;
  exit_code?: number;
  cwd?: string;
  cel_expression?: string;
  cel_vars?: Record<string, any>;
  expected?: any;
  actual?: any;
}
