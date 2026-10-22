// Tests the multipart body a new todo with screenshots is posted as: it carries every field the
// JSON body would, so attaching a screenshot changes nothing else about the todo.
import { describe, expect, it } from 'vitest';
import { todoFormData } from './attachments';

describe('todoFormData', () => {
  const fields = { title: 'Fix it', body: 'Details', priority: 'high', status: 'pending' };

  it('carries a hosted todo\'s visibility', () => {
    expect(todoFormData({ ...fields, visibility: 'public' }, []).get('visibility')).toBe('public');
  });

  it('sends no visibility when none was chosen', () => {
    expect(todoFormData(fields, []).has('visibility')).toBe(false);
  });
});
