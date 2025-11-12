<script>
  import {LoadJSON, SaveChanges} from '../wailsjs/go/main/App.js'
  import FileTree from './lib/FileTree.svelte'
  import {buildFileTree, flattenTree} from './lib/fileTreeUtils.js'

  let files = [];
  let stagedFiles = [];
  let error = '';

  let originalFlatFiles = [];

  async function loadJSON() {
    try {
      const jsonData = await LoadJSON();
      originalFlatFiles = JSON.parse(jsonData);
      files = buildFileTree(originalFlatFiles);
      stagedFiles = JSON.parse(JSON.stringify(files)); // Deep copy
      error = '';
    } catch (e) {
      error = e;
    }
  }

  function dateStamp(nodes) {
    nodes.forEach(node => {
      const date = new Date(node.mod_time);
      const year = date.getFullYear();
      const month = String(date.getMonth() + 1).padStart(2, '0');
      const day = String(date.getDate()).padStart(2, '0');
      const prefix = `${year}${month}${day}`;

      if (!node.name.startsWith(prefix)) {
        node.name = `${prefix}_${node.name}`;
      }
      if (node.children) {
        dateStamp(node.children);
      }
    });
    stagedFiles = stagedFiles; // Trigger reactivity
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

</script>

<main>
  <div class="toolbar">
    <button on:click={loadJSON}>Load JSON</button>
    <button on:click={() => dateStamp(stagedFiles)}>Date Stamp</button>
    <button on:click={saveChanges}>Save Changes</button>
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
      <FileTree files={stagedFiles} editable={true} />
    </div>
  </div>
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
