import { describe, expect, it } from 'vitest';
import { todoFormData } from './attachments';
import { CREATE_RUNTIME } from './createRuntimeTestHelpers';

describe('todoFormData runtime settings', () => {
  const fields = { title: 'Repair parser', body: 'Handle empty input', priority: 'medium', status: 'draft' };

  it('encodes runtime settings and ordered presets alongside attachments', () => {
    const image = { blob: new Blob(['image bytes'], { type: 'image/png' }), name: 'failure.png' };
    const form = todoFormData({ ...fields, step: 'plan', spec: CREATE_RUNTIME, presets: ['reviewer', 'economical'] }, [image]);
    expect(form.get('step')).toBe('plan');
    expect(JSON.parse(String(form.get('spec')))).toEqual(CREATE_RUNTIME);
    expect(JSON.parse(String(form.get('presets')))).toEqual(['reviewer', 'economical']);
    expect((form.get('attachment') as File).name).toBe(image.name);
  });

  it('omits runtime fields for None and preserves an explicitly empty preset list', () => {
    expect(Array.from(todoFormData(fields, []).keys())).toEqual(['title', 'body', 'priority', 'status']);
    expect(todoFormData({ ...fields, step: 'plan', spec: {}, presets: [] }, []).get('presets')).toBe('[]');
  });
});
