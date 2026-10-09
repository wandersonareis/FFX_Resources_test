import { describe, expect, it } from 'bun:test';
import { buildNodeIndex, entryCheckState, entryIdsInNode, type SideNode } from './model';

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

describe('buildNodeIndex', () => {
  it('indexes nested nodes so the keyboard can resolve [data-node-id]', () => {
    const index = buildNodeIndex([imageTree], []);
    expect(index.get('category')).toBe(imageTree);
    expect(index.get('directory')?.id).toBe('directory');
    expect(index.get('leaf:b')?.entry?.id).toBe('b');
    expect(index.get('missing')).toBeUndefined();
  });

  it('merges data/ roots with .vbf roots', () => {
    const vbfRoot: SideNode = { id: 'vbf-root', label: 'mod.vbf', vbf: true, vbfRoot: '/mod.vbf' };
    const index = buildNodeIndex([imageTree], [vbfRoot]);
    expect(index.size).toBe(5);
    expect(index.get('vbf-root')?.vbfRoot).toBe('/mod.vbf');
  });
});
