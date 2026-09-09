import { describe, expect, it } from 'vitest';
import { fixtureFenceSchemasFromDocument } from './fixtureSchema';

describe('fixture schema adapter', () => {
  it('expands fixture fence schemas across aliases', () => {
    const testSchema = { type: 'object' as const, properties: { paths: { type: 'array' } } };
    const lintSchema = { type: 'object' as const, properties: { files: { type: 'array' } } };

    expect(
      fixtureFenceSchemasFromDocument({
        fences: {
          test: { schema: testSchema, aliases: ['yaml test'] },
          lint: { schema: lintSchema, aliases: ['yaml lint'] },
          ignored: { aliases: ['yaml ignored'] },
        },
      }),
    ).toEqual({
      test: testSchema,
      'yaml test': testSchema,
      lint: lintSchema,
      'yaml lint': lintSchema,
    });
  });

  it('returns an empty schema map for malformed documents', () => {
    expect(fixtureFenceSchemasFromDocument(null)).toEqual({});
    expect(fixtureFenceSchemasFromDocument({ fences: [] })).toEqual({});
  });
});
