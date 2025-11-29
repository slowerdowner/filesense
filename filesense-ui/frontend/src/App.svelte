<script>
  import { ScanDirectory, ApplyChanges } from "../wailsjs/go/main/App.js";
  import FileTree from "./lib/FileTree.svelte";
  import RuleEditor from "./lib/RuleEditor.svelte";
  import { buildFileTree, flattenTree } from "./lib/fileTreeUtils.js";
  import { applyRules } from "./lib/ruleEngine.js";

  let files = [];
  let stagedFiles = [];
  let error = "";
  let rules = [];

  let originalFlatFiles = [];

  async function scanDirectory() {
    try {
      originalFlatFiles = await ScanDirectory();
      files = buildFileTree(originalFlatFiles);
      updateStagedFiles();
      error = "";
    } catch (e) {
      error = e;
    }
  }

  function updateStagedFiles() {
    if (files.length === 0) return;
    stagedFiles = applyRules(files, rules);
  }

  function handleRulesChange(event) {
    rules = event.detail;
    updateStagedFiles();
  }

  async function applyChanges() {
    try {
      const flattenedStaged = flattenTree(stagedFiles);
      const changes = flattenedStaged
        .map((file) => {
          const originalFile = originalFlatFiles.find(
            (f) => f.path === file.path,
          );
          if (originalFile && file.name !== originalFile.name) {
            return {
              ...originalFile,
              new_name: file.name,
              new_path:
                file.path.substring(0, file.path.lastIndexOf("/") + 1) +
                file.name,
            };
          }
          return null;
        })
        .filter(Boolean);

      await ApplyChanges(JSON.stringify(changes, null, 2));
      error = "";
      // Refresh the view after applying changes
      await scanDirectory();
    } catch (e) {
      error = e;
    }
  }
</script>

<main>
  <div class="toolbar">
    <button on:click={scanDirectory}>Scan Directory</button>
    <button on:click={applyChanges}>Apply Changes</button>
  </div>
  <div class="container">
    <div class="pane" id="left-pane">
      <h2>Original</h2>
      {#if error}
        <p class="error">{error}</p>
      {/if}
      <FileTree {files} />
    </div>
    <div class="pane" id="center-pane">
      <h2>Rules</h2>
      <RuleEditor {rules} on:change={handleRulesChange} />
    </div>
    <div class="pane" id="right-pane">
      <h2>Preview</h2>
      <FileTree files={stagedFiles} editable={false} />
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
    height: 100vh;
    width: 100vw;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    background: transparent; /* Handled by body */
  }

  .toolbar {
    padding: 1rem 1.5rem;
    background: rgba(15, 23, 42, 0.6);
    border-bottom: var(--pane-border);
    backdrop-filter: blur(10px);
    display: flex;
    gap: 1rem;
    align-items: center;
  }

  button {
    background: var(--accent-color);
    color: white;
    padding: 8px 16px;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    font-weight: 500;
    font-size: 0.9rem;
    transition: all 0.2s;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  button:hover {
    background: var(--accent-hover);
    transform: translateY(-1px);
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.15);
  }

  .container {
    display: grid;
    grid-template-columns: 1fr 350px 1fr;
    gap: 1.5rem;
    padding: 1.5rem;
    flex-grow: 1;
    overflow: hidden;
  }

  .pane {
    background: var(--pane-bg);
    border-radius: var(--radius);
    padding: 1.5rem;
    border: var(--pane-border);
    backdrop-filter: blur(10px);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    box-shadow: var(--shadow);
  }

  h2 {
    margin: 0 0 1rem 0;
    padding-bottom: 0.8rem;
    border-bottom: var(--pane-border);
    font-size: 1.1rem;
    font-weight: 600;
    color: var(--text-primary);
    letter-spacing: 0.02em;
  }

  .error {
    color: var(--danger-color);
    background: rgba(239, 68, 68, 0.1);
    padding: 0.8rem;
    border-radius: 6px;
    margin-bottom: 1rem;
    border: 1px solid rgba(239, 68, 68, 0.2);
    font-size: 0.9rem;
  }
</style>
