<script>
  import {LoadJSON, SaveChanges, ZipFiles, SelectZipFile} from '../wailsjs/go/main/App.js'
  import FileTree from './lib/FileTree.svelte'
  import ContextMenu from './lib/ContextMenu.svelte'
  import {buildFileTree, flattenTree} from './lib/fileTreeUtils.js'
  import { setContext } from 'svelte';
  import { selectedPaths } from './lib/stores.js';

  let files = [];
  let stagedFiles = [];
  let error = '';

  let originalFlatFiles = [];
  let lastSelected = null;

  let history = [];

  let contextMenu = {
    visible: false,
    x: 0,
    y: 0,
    file: null,
  };

  setContext('selection', {
    setLastSelected: (path) => lastSelected = path,
    getLastSelected: () => lastSelected,
  });

  async function loadJSON() {
    try {
      const jsonData = await LoadJSON();
      originalFlatFiles = JSON.parse(jsonData);
      files = buildFileTree(originalFlatFiles);
      stagedFiles = JSON.parse(JSON.stringify(files)); // Deep copy
      history = [JSON.stringify(stagedFiles)];
      error = '';
    } catch (e) {
      error = e;
    }
  }

  function pushState() {
    history.push(JSON.stringify(stagedFiles));
    history = history;
  }

  function dateStamp(nodes) {
    pushState();
    let currentSelectedPaths;
    selectedPaths.subscribe(value => currentSelectedPaths = value)();

    function traverse(nodes) {
      nodes.forEach(node => {
        if (currentSelectedPaths.has(node.path)) {
          const date = new Date(node.mod_time);
          const year = date.getFullYear();
          const month = String(date.getMonth() + 1).padStart(2, '0');
          const day = String(date.getDate()).padStart(2, '0');
          const prefix = `${year}${month}${day}`;

          if (!node.name.startsWith(prefix)) {
            node.name = `${prefix}_${node.name}`;
          }
        }
        if (node.children) {
          traverse(node.children);
        }
      });
    }
    traverse(nodes);
    stagedFiles = stagedFiles; // Trigger reactivity
  }

  function startOver() {
    stagedFiles = JSON.parse(JSON.stringify(files));
    history = [JSON.stringify(stagedFiles)];
  }

  function undo() {
    if (history.length > 1) {
      history.pop();
      stagedFiles = JSON.parse(history[history.length - 1]);
      history = history;
    }
  }


  async function saveChanges() {
    try {
      const flattenedStaged = flattenTree(stagedFiles);
      const changes = flattenedStaged.map((file) => {
        const originalFile = originalFlatFiles.find(f => f.path === file.path);
        if (originalFile && file.name !== originalFile.name) {
          return {
            ...originalFile,
            new_name: file.name,
            new_path: file.path.substring(0, file.path.lastIndexOf('/') + 1) + file.name,
          };
        }
        return null;
      }).filter(Boolean);

      await SaveChanges(JSON.stringify(changes, null, 2));
      error = '';
    } catch (e) {
      error = e;
    }
  }

  function showContextMenu(event) {
    contextMenu.visible = true;
    contextMenu.x = event.detail.x;
    contextMenu.y = event.detail.y;
    contextMenu.file = event.detail.file;
  }

  function handleContextMenuAction(event) {
    contextMenu.visible = false;
    const action = event.detail;
    if (action === 'delete') {
      softDelete();
    } else if (action === 'zip') {
      zipSelected();
    }
  }

  function closeContextMenu() {
    contextMenu.visible = false;
  }

  function handleMove(event) {
    pushState();
    const { draggedFilePath, targetPath } = event.detail;

    let draggedFile, targetFile, draggedFileParent;

    function findFile(nodes, path, parent = null) {
      for (let i = 0; i < nodes.length; i++) {
        const node = nodes[i];
        if (node.path === path) {
          return {node, parent};
        }
        if (node.children) {
          const found = findFile(node.children, path, node);
          if (found) return found;
        }
      }
      return null;
    }

    const dragged = findFile(stagedFiles, draggedFilePath);
    draggedFile = dragged.node;
    draggedFileParent = dragged.parent;

    const target = findFile(stagedFiles, targetPath);
    targetFile = target.node;

    if (!draggedFile || !targetFile) return;

    // Remove from old parent
    const sourceList = draggedFileParent ? draggedFileParent.children : stagedFiles;
    const index = sourceList.findIndex(f => f.path === draggedFilePath);
    sourceList.splice(index, 1);

    // Add to new parent
    if (targetFile.is_dir) {
      targetFile.children.push(draggedFile);
      draggedFile.path = targetFile.path + '/' + draggedFile.name;
    } else {
      const targetParent = findFile(stagedFiles, targetPath).parent;
      const destList = targetParent ? targetParent.children : stagedFiles;
      destList.push(draggedFile);
      draggedFile.path = (targetParent ? targetParent.path + '/' : '') + draggedFile.name;
    }

    stagedFiles = stagedFiles;
  }

  function softDelete() {
    pushState();
    let currentSelectedPaths;
    selectedPaths.subscribe(value => currentSelectedPaths = value)();

    let discardFolder = stagedFiles.find(f => f.name === '99_Discard' && f.is_dir);
    if (!discardFolder) {
      discardFolder = {
        name: '99_Discard',
        path: '99_Discard',
        is_dir: true,
        children: [],
        collapsed: false,
      };
      stagedFiles.push(discardFolder);
    }

    let filesToMove = [];

    function findFilesToMove(nodes) {
      for (let i = nodes.length - 1; i >= 0; i--) {
        const node = nodes[i];
        if (currentSelectedPaths.has(node.path)) {
          filesToMove.push(node);
          nodes.splice(i, 1);
        } else if (node.children) {
          findFilesToMove(node.children);
        }
      }
    }

    findFilesToMove(stagedFiles);

    filesToMove.forEach(file => {
      discardFolder.children.push(file);
      file.path = discardFolder.path + '/' + file.name;
    });

    stagedFiles = stagedFiles;
  }

  async function zipSelected() {
    let currentSelectedPaths;
    selectedPaths.subscribe(value => currentSelectedPaths = value)();

    const paths = Array.from(currentSelectedPaths);
    if (paths.length === 0) return;

    try {
      const dest = await SelectZipFile();
      if (dest) {
        await ZipFiles(paths, dest);
      }
    } catch (e) {
      error = e;
    }
  }

</script>

<main on:click={closeContextMenu}>
  <div class="toolbar">
    <button on:click={loadJSON}>Load JSON</button>
    <button on:click={() => dateStamp(stagedFiles)}>Date Stamp</button>
    <button on:click={saveChanges}>Save Changes</button>
    <button on:click={startOver}>Start Over</button>
    <button on:click={undo} disabled={history.length <= 1}>Undo</button>
  </div>
  <div class="container">
    <div class="pane" id="left-pane">
      <h2>Original</h2>
      {#if error}
        <p class="error">{error}</p>
      {/if}
      <FileTree files={files} />
    </div>
    <div class="pane" id="right-pane">
      <h2>Changes</h2>
      <FileTree files={stagedFiles} editable={true} on:contextmenu={showContextMenu} on:move={handleMove} />
    </div>
  </div>

  {#if contextMenu.visible}
    <ContextMenu x={contextMenu.x} y={contextMenu.y} on:action={handleContextMenuAction} />
  {/if}
</main>

<style>
  :root {
    --background-color: #1e1e1e;
    --pane-background: #2a2a2a;
    --text-color: #e0e0e0;
    --border-color: #444;
    --button-bg: #4caf50;
    --button-hover-bg: #45a049;
    --error-color: #f44336;
  }

  main {
    background-color: var(--background-color);
    color: var(--text-color);
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol";
    height: 100vh;
    width: 100vw;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
  }

  .toolbar {
    padding: 0.5rem;
    background-color: var(--pane-background);
    border-bottom: 1px solid var(--border-color);
  }

  button {
    background-color: var(--button-bg);
    color: white;
    padding: 10px 15px;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    margin-right: 0.5rem;
  }

  button:hover {
    background-color: var(--button-hover-bg);
  }

  button:disabled {
    background-color: #555;
    cursor: not-allowed;
  }

  .container {
    display: flex;
    flex-grow: 1;
    padding: 1rem;
    box-sizing: border-box;
    overflow-y: hidden;
  }

  .pane {
    flex: 1;
    background-color: var(--pane-background);
    border-radius: 8px;
    padding: 1rem;
    margin: 0 0.5rem;
    border: 1px solid var(--border-color);
    overflow-y: auto;
  }

  h2 {
    margin-top: 0;
    border-bottom: 1px solid var(--border-color);
    padding-bottom: 0.5rem;
  }

  .error {
    color: var(--error-color);
  }
</style>
