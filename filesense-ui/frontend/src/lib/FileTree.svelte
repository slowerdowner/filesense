<script>
  export let files = [];
  export let editable = false;

  function handleNameChange(event, file) {
    file.name = event.target.innerText;
  }
</script>

<ul>
  {#each files as file}
    <li>
      <span
        class:folder={file.is_dir}
        contenteditable={editable}
        on:blur={(event) => handleNameChange(event, file)}
        on:keydown={(event) => { if (event.key === 'Enter') { event.target.blur(); } }}
      >{file.name}</span>

      {#if file.children && file.children.length > 0}
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
    top: 0.6em;
    width: 1rem;
    height: 1px;
    background-color: #444;
  }


  span {
    padding: 2px 5px;
    border-radius: 3px;
  }

  .folder {
    font-weight: bold;
  }

  span[contenteditable="true"]:focus {
    background-color: #555;
    outline: 2px solid #777;
  }
</style>
