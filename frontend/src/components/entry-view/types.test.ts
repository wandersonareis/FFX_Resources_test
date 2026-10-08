import { describe, expect, it } from 'vitest';
import { entryCheckState, entryIdsInNode, type SideNode } from './types';

const imageTree: SideNode = {
  id: 'category',
  label: 'category',
  kind: 'images',
  children: [
    {
      id: 'directory',
      label: 'directory',
      kind: 'images',
      children: [
        {
          id: 'leaf:a',
          label: 'a',
          kind: 'images',
          entry: { kind: 'images', id: 'a', key: 'a', label: 'a' },
        },
        {
          id: 'leaf:b',
          label: 'b',
          kind: 'images',
          entry: { kind: 'images', id: 'b', key: 'b', label: 'b' },
        },
      ],
    },
  ],
};

describe('data tree checkbox selection', () => {
  it('collects every image leaf under nested groups', () => {
    expect(entryIdsInNode(imageTree)).toEqual(['a', 'b']);
  });

  it('calculates tri-state from every descendant leaf', () => {
    expect(entryCheckState(imageTree, new Set())).toBe(false);
    expect(entryCheckState(imageTree, new Set(['a']))).toBe('indeterminate');
    expect(entryCheckState(imageTree, new Set(['a', 'b']))).toBe(true);
  });
});
