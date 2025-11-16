export function buildFileTree(files) {
  const tree = [];
  const map = new Map();

  files.forEach(file => {
    const newNode = { ...file, children: [] };
    if (newNode.is_dir) {
      newNode.collapsed = true;
    }
    map.set(file.path, newNode);
  });

  files.forEach(file => {
    const parentPath = file.path.substring(0, file.path.lastIndexOf('/'));
    if (parentPath && map.has(parentPath)) {
      map.get(parentPath).children.push(map.get(file.path));
    } else {
      tree.push(map.get(file.path));
    }
  });

  return tree;
}

export function flattenTree(nodes) {
  const flat = [];
  function traverse(node) {
    const { children, ...rest } = node;
    flat.push(rest);
    if (children) {
      children.forEach(traverse);
    }
  }
  nodes.forEach(traverse);
  return flat;
}
