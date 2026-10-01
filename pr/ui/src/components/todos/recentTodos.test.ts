import { beforeEach, describe, expect, it } from 'vitest';
import { loadRecentTodos, RECENT_TODOS_LIMIT, recentTodoHost, rememberRecentTodo, type RecentTodo } from './recentTodos';

const appHost = 'localhost:5173';
const otherHost = 'app.example.com';

function entry(ref: string, dir = '/work/acme', at = 1): RecentTodo {
  return { ref, title: `Todo ${ref}`, dir, at };
}

beforeEach(() => localStorage.clear());

describe('recentTodoHost', () => {
  it('uses the source page host, including its port', () => {
    expect(recentTodoHost(`http://${appHost}/checkout?step=2`)).toBe(appHost);
  });

  it('falls back to gavel\'s own host when there is no usable source url', () => {
    expect(recentTodoHost('')).toBe(window.location.host);
    expect(recentTodoHost('not a url')).toBe(window.location.host);
  });
});

describe('rememberRecentTodo', () => {
  it('keeps each host\'s todos separate', () => {
    rememberRecentTodo(appHost, entry('a'));
    rememberRecentTodo(otherHost, entry('b'));

    expect(loadRecentTodos(appHost).map(e => e.ref)).toEqual(['a']);
    expect(loadRecentTodos(otherHost).map(e => e.ref)).toEqual(['b']);
  });

  it('moves a re-used todo to the front instead of duplicating it', () => {
    rememberRecentTodo(appHost, entry('a'));
    rememberRecentTodo(appHost, entry('b'));
    rememberRecentTodo(appHost, entry('a', '/work/acme', 3));

    expect(loadRecentTodos(appHost).map(e => [e.ref, e.at])).toEqual([['a', 3], ['b', 1]]);
  });

  it('treats the same ref in another workspace as a different todo', () => {
    rememberRecentTodo(appHost, entry('a', '/work/acme'));
    rememberRecentTodo(appHost, entry('a', '/work/other'));

    expect(loadRecentTodos(appHost).map(e => e.dir)).toEqual(['/work/other', '/work/acme']);
  });

  it(`keeps only the newest ${RECENT_TODOS_LIMIT}`, () => {
    const refs = Array.from({ length: RECENT_TODOS_LIMIT + 2 }, (_, i) => `t${i}`);
    refs.forEach(ref => rememberRecentTodo(appHost, entry(ref)));

    expect(loadRecentTodos(appHost).map(e => e.ref)).toEqual(refs.slice(2).reverse());
  });

  it('starts empty when the stored value is corrupt', () => {
    localStorage.setItem('gavel.pr-ui.todoNew.recent.v1', '{not json');
    expect(loadRecentTodos(appHost)).toEqual([]);
  });
});
