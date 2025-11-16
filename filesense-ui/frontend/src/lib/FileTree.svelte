<script>
  import { selectedPaths } from './stores.js';
  import { getContext, onMount, createEventDispatcher } from 'svelte';

  export let files = [];
  export let editable = false;

  let lastSelected = null;
  const dispatch = createEventDispatcher();

  // Use a context to share the last selected item across all FileTree components
  const { setLastSelected, getLastSelected } = getContext('selection');

  onMount(() => {
    lastSelected = getLastSelected();
  });


  function handleNameChange(event, file) {
    file.name = event.target.innerText;
  }

  function toggleNode(file, event) {
    if (file.is_dir) {
      file.collapsed = !file.collapsed;
    }
    handleSelect(file, event);
  }

  function handleSelect(file, event) {
    let currentSelectedPaths;
    selectedPaths.subscribe(value => currentSelectedPaths = value)();

    if (event.shiftKey && lastSelected) {
      const allPaths = flattenPaths(files);
      const lastIndex = allPaths.indexOf(lastSelected);
      const currentIndex = allPaths.indexOf(file.path);
      const [start, end] = [lastIndex, currentIndex].sort((a, b) => a - b);
      const newSelection = new Set(currentSelectedPaths);
      for (let i = start; i <= end; i++) {
        newSelection.add(allPaths[i]);
      }
      selectedPaths.set(newSelection);
    } else if (event.ctrlKey || event.metaKey) {
      selectedPaths.update(paths => {
        paths.has(file.path) ? paths.delete(file.path) : paths.add(file.path);
        return paths;
      });
    } else {
      selectedPaths.set(new Set([file.path]));
    }
    setLastSelected(file.path);
  }

  function flattenPaths(nodes) {
    let paths = [];
    nodes.forEach(node => {
      paths.push(node.path);
      if (node.children && !node.collapsed) {
        paths = paths.concat(flattenPaths(node.children));
      }
    });
    return paths;
  }

  function handleCheckbox(file, event) {
    selectedPaths.update(paths => {
      event.target.checked ? paths.add(file.path) : paths.delete(file.path);
      return paths;
    });
  }

  function showContextMenu(event, file) {
    event.preventDefault();
    dispatch('contextmenu', { x: event.clientX, y: event.clientY, file });
  }

  // Drag and Drop
  let dragCounter = 0;

  function handleDragStart(event, file) {
    event.dataTransfer.setData('text/plain', file.path);
  }

  function handleDragOver(event) {
    event.preventDefault();
    event.currentTarget.classList.add('drag-over');
    dragCounter++;
  }

  function handleDragLeave(event) {
    dragCounter--;
    if (dragCounter === 0) {
      event.currentTarget.classList.remove('drag-over');
    }
  }

  function handleDrop(event, targetFile) {
    event.preventDefault();
    dragCounter = 0;
    event.currentTarget.classList.remove('drag-over');
    const draggedFilePath = event.dataTransfer.getData('text/plain');
    dispatch('move', { draggedFilePath, targetPath: targetFile.path });
  }

</script>

<ul>
  {#each files as file}
    <li>
      <div
        class="node-content"
        class:selected={$selectedPaths.has(file.path)}
        on:click={(event) => toggleNode(file, event)}
        on:contextmenu={(event) => showContextMenu(event, file)}
        draggable={editable}
        on:dragstart={(event) => handleDragStart(event, file)}
        on:dragover={handleDragOver}
        on:dragleave={handleDragLeave}
        on:drop={(event) => handleDrop(event, file)}
      >
        <input type="checkbox" checked={$selectedPaths.has(file.path)} on:click|stopPropagation={(e) => handleCheckbox(file, e)} />
        {#if file.is_dir}
          <span class="indicator">{file.collapsed ? '▶' : '▼'}</span>
        {/if}
        <span
          class="name"
          class:folder={file.is_dir}
          contenteditable={editable}
          on:click|stopPropagation={(event) => handleSelect(file, event)}
          on:blur={(event) => handleNameChange(event, file)}
          on:keydown={(event) => { if (event.key === 'Enter') { event.target.blur(); } }}
        >{file.name}</span>
      </div>

      {#if file.children && file.children.length > 0 && !file.collapsed}
        <svelte:self files={file.children} {editable} />
      {/if}
    </li>
  {/each}
</ul>

<style>
  ul {
    list-style-type: none;
    padding-left: 1.5rem;
    border-left: 1px solid #444;
  }

  li {
    margin: 0.25rem 0;
    position: relative;
  }

  li::before {
    content: '';
    position: absolute;
    left: -1.5rem;
    top: 0.8em;
    width: 1rem;
    height: 1px;
    background-color: #444;
  }

  .node-content {
    display: flex;
    align-items: center;
    cursor: pointer;
    border-radius: 3px;
  }

  .node-content.selected {
    background-color: #0078d4;
  }

  .drag-over {
    background-color: #0078d4;
    opacity: 0.5;
  }

  input[type="checkbox"] {
    margin-right: 5px;
  }

  .indicator {
    width: 1em;
    margin-right: 5px;
    user-select: none;
  }

  .name {
    padding: 2px 5px;
    border-radius: 3px;
    cursor: text;
  }

  .folder {
    font-weight: bold;
    cursor: pointer;
  }

  .name[contenteditable="true"]:focus {
    background-color: #555;
    outline: 2px solid #777;
  }
</style>
