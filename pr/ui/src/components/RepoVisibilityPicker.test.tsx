// Tests the hosted top bar's repository picker: the organization's repositories grouped by owner,
// hiding and showing an owner or one repository, and finding one by name.
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Project } from '../types';
import { RepoVisibilityPicker } from './RepoVisibilityPicker';

function repo(name: string, owner: string, extra: Partial<Project> = {}): Project {
  return { name, dir: `repo:${name}`, repos: [], ...(owner ? { owner } : {}), ...extra };
}

const projects = [
  repo('acme/api', 'github.com/acme'),
  repo('acme/web', 'github.com/acme', { hidden: true }),
  repo('Platform/rules', 'dev.azure.com/contoso', { hidden: true, ownerHidden: true }),
  repo('scratch', ''),
];

function renderPicker() {
  const onSetRepoHidden = vi.fn();
  const onSetOwnerHidden = vi.fn();
  render(<RepoVisibilityPicker projects={projects} onSetRepoHidden={onSetRepoHidden} onSetOwnerHidden={onSetOwnerHidden} />);
  fireEvent.click(screen.getByRole('button', { name: /2 of 4 repositories/ }));
  return { onSetRepoHidden, onSetOwnerHidden };
}

describe('RepoVisibilityPicker', () => {
  it('says how many repositories show, and groups them by owner, pushed ones apart', () => {
    renderPicker();
    expect(screen.getByText('github.com/acme')).toBeTruthy();
    expect(screen.getByText('dev.azure.com/contoso')).toBeTruthy();
    expect(screen.getByText('Pushed repositories')).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Hide Pushed repositories' })).toBeNull();
  });

  it("hides and shows a whole owner, including repositories it gains later", () => {
    const { onSetOwnerHidden } = renderPicker();
    fireEvent.click(screen.getByRole('button', { name: 'Hide github.com/acme' }));
    fireEvent.click(screen.getByRole('button', { name: 'Show dev.azure.com/contoso' }));
    expect(onSetOwnerHidden.mock.calls).toEqual([['github.com/acme', true], ['dev.azure.com/contoso', false]]);
  });

  it('hides and shows one repository, but not one its hidden owner hides', () => {
    const { onSetRepoHidden } = renderPicker();
    fireEvent.click(screen.getByRole('button', { name: 'Hide acme/api' }));
    fireEvent.click(screen.getByRole('button', { name: 'Show acme/web' }));
    expect(onSetRepoHidden.mock.calls).toEqual([['repo:acme/api', true], ['repo:acme/web', false]]);
    const rules = screen.getByRole('button', { name: 'Show Platform/rules' }) as HTMLButtonElement;
    expect(rules.disabled).toBe(true);
    expect(rules.title).toBe('Hidden with dev.azure.com/contoso');
  });

  it('finds a repository by name', () => {
    renderPicker();
    fireEvent.change(screen.getByPlaceholderText('Find a repository'), { target: { value: 'rul' } });
    expect(screen.getByText('Platform/rules')).toBeTruthy();
    expect(screen.queryByText('acme/api')).toBeNull();
    expect(screen.queryByText('github.com/acme')).toBeNull();
  });

  it('says all repositories show when none is hidden', () => {
    render(<RepoVisibilityPicker projects={[repo('acme/api', 'github.com/acme')]} onSetRepoHidden={vi.fn()} onSetOwnerHidden={vi.fn()} />);
    expect(screen.getByRole('button', { name: /All repositories/ })).toBeTruthy();
  });
});
