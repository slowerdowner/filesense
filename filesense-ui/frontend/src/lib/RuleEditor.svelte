<script>
    import { createEventDispatcher } from "svelte";

    export let rules = [];

    const dispatch = createEventDispatcher();

    function addRule(type) {
        const newRule = {
            id: Date.now(),
            type,
            enabled: true,
            params: {},
        };

        // Set default params based on type
        switch (type) {
            case "replace":
                newRule.params = { find: "", replace: "" };
                break;
            case "regex":
                newRule.params = { pattern: "", replace: "", flags: "g" };
                break;
            case "prefix":
                newRule.params = { text: "" };
                break;
            case "suffix":
                newRule.params = { text: "" };
                break;
            case "extension":
                newRule.params = { mode: "lowercase" };
                break;
        }

        rules = [...rules, newRule];
        dispatch("change", rules);
    }

    function removeRule(id) {
        rules = rules.filter((r) => r.id !== id);
        dispatch("change", rules);
    }

    function updateRule() {
        dispatch("change", rules);
    }
</script>

<div class="rule-editor">
    <div class="controls">
        <button on:click={() => addRule("replace")}>+ Replace</button>
        <button on:click={() => addRule("regex")}>+ Regex</button>
        <button on:click={() => addRule("prefix")}>+ Prefix</button>
        <button on:click={() => addRule("suffix")}>+ Suffix</button>
        <button on:click={() => addRule("extension")}>+ Extension</button>
        <button on:click={() => addRule("date_prefix")}>+ Date Prefix</button>
    </div>

    <div class="rules-list">
        {#each rules as rule (rule.id)}
            <div class="rule-item">
                <div class="rule-header">
                    <input
                        type="checkbox"
                        bind:checked={rule.enabled}
                        on:change={updateRule}
                    />
                    <span class="rule-type">{rule.type}</span>
                    <button
                        class="remove-btn"
                        on:click={() => removeRule(rule.id)}>×</button
                    >
                </div>

                <div class="rule-params">
                    {#if rule.type === "replace"}
                        <input
                            placeholder="Find"
                            bind:value={rule.params.find}
                            on:input={updateRule}
                        />
                        <input
                            placeholder="Replace"
                            bind:value={rule.params.replace}
                            on:input={updateRule}
                        />
                    {:else if rule.type === "regex"}
                        <input
                            placeholder="Pattern"
                            bind:value={rule.params.pattern}
                            on:input={updateRule}
                        />
                        <input
                            placeholder="Replace"
                            bind:value={rule.params.replace}
                            on:input={updateRule}
                        />
                    {:else if rule.type === "prefix"}
                        <input
                            placeholder="Prefix text"
                            bind:value={rule.params.text}
                            on:input={updateRule}
                        />
                    {:else if rule.type === "suffix"}
                        <input
                            placeholder="Suffix text"
                            bind:value={rule.params.text}
                            on:input={updateRule}
                        />
                    {:else if rule.type === "extension"}
                        <select
                            bind:value={rule.params.mode}
                            on:change={updateRule}
                        >
                            <option value="lowercase">Lowercase</option>
                            <option value="uppercase">Uppercase</option>
                        </select>
                    {:else if rule.type === "date_prefix"}
                        <span class="info">Adds YYYYMMDD_ prefix</span>
                    {/if}
                </div>
            </div>
        {/each}
    </div>
</div>

<style>
    .rule-editor {
        background: var(--pane-bg);
        padding: 1rem;
        border-radius: var(--radius);
        margin-bottom: 1rem;
        border: var(--pane-border);
        backdrop-filter: blur(10px);
    }

    .controls {
        display: flex;
        gap: 0.5rem;
        margin-bottom: 1rem;
        flex-wrap: wrap;
    }

    button {
        background: rgba(255, 255, 255, 0.1);
        border: 1px solid rgba(255, 255, 255, 0.2);
        color: var(--text-primary);
        padding: 6px 12px;
        border-radius: 6px;
        cursor: pointer;
        transition: all 0.2s;
        font-size: 0.9rem;
    }

    button:hover {
        background: rgba(255, 255, 255, 0.2);
        transform: translateY(-1px);
    }

    .rules-list {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }

    .rule-item {
        background: rgba(0, 0, 0, 0.2);
        padding: 0.8rem;
        border-radius: 8px;
        border: 1px solid rgba(255, 255, 255, 0.05);
        transition: border-color 0.2s;
    }

    .rule-item:hover {
        border-color: rgba(255, 255, 255, 0.1);
    }

    .rule-header {
        display: flex;
        align-items: center;
        gap: 0.8rem;
        margin-bottom: 0.8rem;
    }

    .rule-type {
        font-weight: 600;
        text-transform: uppercase;
        font-size: 0.75rem;
        letter-spacing: 0.05em;
        color: var(--accent-color);
        flex-grow: 1;
    }

    .remove-btn {
        background: transparent;
        color: var(--text-secondary);
        border: none;
        font-size: 1.2rem;
        padding: 0 0.5rem;
        cursor: pointer;
        opacity: 0.7;
    }

    .remove-btn:hover {
        color: var(--danger-color);
        opacity: 1;
        background: transparent;
        transform: none;
    }

    .rule-params {
        display: flex;
        gap: 0.5rem;
    }

    input,
    select {
        background: rgba(0, 0, 0, 0.3);
        border: 1px solid rgba(255, 255, 255, 0.1);
        color: var(--text-primary);
        padding: 8px 12px;
        border-radius: 6px;
        flex: 1;
        font-size: 0.9rem;
        outline: none;
        transition: border-color 0.2s;
    }

    input:focus,
    select:focus {
        border-color: var(--accent-color);
    }

    .info {
        font-size: 0.85rem;
        color: var(--text-secondary);
        font-style: italic;
    }
</style>
