import type { FixtureFenceSchemas } from '@flanksource/clicky-ui/data';

type FixtureSchemaDocument = {
  fences?: Record<
    string,
    {
      schema?: FixtureFenceSchemas[string];
      aliases?: string[];
    }
  >;
};

export function fixtureFenceSchemasFromDocument(doc: unknown): FixtureFenceSchemas {
  if (!isFixtureSchemaDocument(doc)) return {};

  const schemas: FixtureFenceSchemas = {};
  for (const [name, config] of Object.entries(doc.fences ?? {})) {
    if (!isSchemaObject(config.schema)) continue;
    schemas[name] = config.schema;
    for (const alias of config.aliases ?? []) {
      if (typeof alias === 'string' && alias.trim()) {
        schemas[alias] = config.schema;
      }
    }
  }
  return schemas;
}

function isFixtureSchemaDocument(value: unknown): value is FixtureSchemaDocument {
  return typeof value === 'object' && value != null && !Array.isArray(value);
}

function isSchemaObject(value: unknown): value is FixtureFenceSchemas[string] {
  return typeof value === 'object' && value != null && !Array.isArray(value);
}
