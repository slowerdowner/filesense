<script>
  import { createEventDispatcher } from 'svelte';

  export let x;
  export let y;

  const dispatch = createEventDispatcher();

  function handleAction(action) {
    dispatch('action', action);
  }
</script>

<div class="context-menu" style="left: {x}px; top: {y}px;">
  <ul>
    <li on:click={() => handleAction('rename')}>Rename</li>
    <li on:click={() => handleAction('delete')}>Delete</li>
    <li on:click={() => handleAction('zip')}>Zip</li>
  </ul>
</div>

<style>
  .context-menu {
    position: absolute;
    background-color: #3a3a3a;
    border: 1px solid #555;
    border-radius: 4px;
    padding: 5px 0;
    z-index: 1000;
  }

  ul {
    list-style: none;
    margin: 0;
    padding: 0;
  }

  li {
    padding: 8px 15px;
    cursor: pointer;
  }

  li:hover {
    background-color: #555;
  }
</style>
