import { describe, expect, it } from 'vitest';
import { layerOfOrigin, provenanceForPath, type GavelTrace } from './provenance';

const trace: GavelTrace = {
  sources: [
    { origin: 'user-home', path: '~/.gavel.yaml', config: { verify: { model: 'gemini' }, commit: {} } },
    {
      origin: 'target-directory',
      path: '/repo/.gavel.yaml',
      config: { verify: { model: 'claude' }, commit: { model: 'opus' }, lint: { ignore: [] } },
    },
  ],
  merged: {},
};

describe('layerOfOrigin', () => {
  it('maps user-home to user and repo origins to project', () => {
    expect(layerOfOrigin('user-home')).toBe('user');
    expect(layerOfOrigin('git-root')).toBe('project');
    expect(layerOfOrigin('target-directory')).toBe('project');
  });
});

describe('provenanceForPath', () => {
  it('gives the highest-priority layer that sets a field (later sources win)', () => {
    // verify.model is set in both layers; the repo layer is applied last.
    expect(provenanceForPath(trace, 'verify.model')).toBe('project');
  });

  it('attributes a field set only in the user layer to user', () => {
    const userOnly: GavelTrace = {
      sources: [trace.sources![0]],
      merged: {},
    };
    expect(provenanceForPath(userOnly, 'verify.model')).toBe('user');
  });

  it('attributes a field set only in the project layer to project', () => {
    expect(provenanceForPath(trace, 'commit.model')).toBe('project');
  });

  it('returns undefined for a field no layer sets (built-in default)', () => {
    expect(provenanceForPath(trace, 'todos.runPrompt')).toBeUndefined();
  });

  it('treats an empty array / object as unset', () => {
    // lint.ignore is present but empty in the repo layer → not a real override.
    expect(provenanceForPath(trace, 'lint.ignore')).toBeUndefined();
  });

  it('returns undefined when the trace has no sources', () => {
    expect(provenanceForPath(null, 'verify.model')).toBeUndefined();
    expect(provenanceForPath({ merged: {} }, 'verify.model')).toBeUndefined();
  });
});
